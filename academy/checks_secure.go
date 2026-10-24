package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

// Checked exercises for Stage 17 (secure coding): a parser that never
// trusts a length, a secret scanner, and SHA-256 for verifying downloads.
var secureChecks = []Check{
	{
		ID: "tlv_parse", Stage: 17, Concept: "Safe Parsing: Lengths, Limits & Binary Formats",
		Title: "A parser that never trusts a length",
		Task:  "Read binary records from standard input and print them. Each record is a type byte, a two-byte big-endian length L, then L bytes of value. Type 1 is text: print text \"…\" with each byte from space to ~ as itself, except \" and \\, and every other byte as \\xHH (two lowercase hex digits). Type 2 is a number: its length must be exactly 4, and it prints as number N (unsigned, big-endian). Type 3 is a list: its value is itself a sequence of records; print list {, then the records inside, indented two more spaces, then }. Any other type prints skip type T, L bytes. Records at the top level are not indented. Refuse, with a message on standard error, exit status 1 and nothing at all on standard output: input longer than 1,048,576 bytes, a header cut short, a length longer than what remains of the input or of the enclosing list, a number whose length is not 4, and lists nested more than 8 deep. Empty input prints nothing.",
		File:  "tlv_parse.c",
		Starter: `/* Checked exercise: tlv_parse
 *   input bytes:  01 00 02 'h' 'i'   03 00 07   02 00 04 00 00 01 00
 *   output:       text "hi"
 *                 list {
 *                   number 256
 *                 }
 * Every length is a claim by whoever made the file. Before using one,
 * compare it with the bytes that really remain:
 *   if (len > remaining) refuse;   (never: offset + len > size)
 * Read the input into an unsigned char buffer (a plain char is signed:
 * a length byte of 0xC8 would become -56).
 * Validate everything before printing anything: a file that fails on
 * its last record must print nothing. One way: a function
 * parse(buf, n, depth, print) called twice, first with print = 0.
 */
#include <stdio.h>
#include <stdlib.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "\x01\x00\x02hi\x03\x00\x07\x02\x00\x04\x00\x00\x01\x00", Want: "text \"hi\"\nlist {\n  number 256\n}\n"},
			{Name: "every type", Gen: generated(func() (string, string) {
				in := tlvRec(1, "hello") + tlvRec(2, "\x00\x01\x00\x00") + tlvRec(9, "abc") + tlvRec(3, tlvRec(3, "")+tlvRec(1, "")) + tlvRec(2, "\xff\xff\xff\xff")
				return in, tlvWant(in)
			}),
				Explain: "An unknown type is skipped by its length, an empty list prints list { and }, and the largest number is 4294967295: read it as unsigned."},
			{Name: "text that needs escaping", Gen: generated(func() (string, string) {
				in := tlvRec(1, "say \"hi\"\\\n\x00\x1b[31m\xc3\xa9~ ")
				return in, tlvWant(in)
			}),
				Explain: "Quotes, backslashes, control bytes (a newline, a NUL, the escape that starts a terminal colour code) and non-ASCII bytes are printed as \\xHH. Printing them raw would let a file inject terminal commands or fake output lines."},
			{Name: "lengths of 128 and more", Gen: generated(func() (string, string) {
				in := tlvRec(1, strings.Repeat("x", 200)) + tlvRec(7, strings.Repeat("y", 40000))
				return in, tlvWant(in)
			}),
				Explain: "Length bytes of 0xC8 or 0x9C are above 127: read through a plain char, they turn negative and the length becomes garbage. Use unsigned char."},
			{Name: "a length past the end", Stdin: "\x01\x00\x02hi\x01\x00\xc8ping", WantExit: 1, Stderr: true,
				Explain: "The second record claims 200 bytes and 4 follow: Heartbleed's bug. Using the length without comparing it with what remains reads past the end of the input, and cutting it down to what remains is no better: the file is lying, so refuse it."},
			{Name: "a length past the end of its list", Gen: generated(func() (string, string) {
				return "\x03\x00\x05" + "\x01\x00\x09abcdefghi", ""
			}), WantExit: 1, Stderr: true,
				Explain: "The list is 5 bytes long, but the text inside it claims 9: the text would run out of the list into whatever follows. Each list's records must fit inside the list, not just inside the input."},
			{Name: "a header cut short", Stdin: "\x02\x00\x04\x00\x00\x00\x01\x01\x00", WantExit: 1, Stderr: true},
			{Name: "a number of 3 bytes", Stdin: "\x02\x00\x03abc", WantExit: 1, Stderr: true},
			{Name: "an error after valid records", Gen: generated(func() (string, string) {
				return tlvRec(1, "first") + tlvRec(2, "\x00\x00\x00\x07") + "\x01\x00\x10short", ""
			}), WantExit: 1, Stderr: true,
				Explain: "The first two records are fine and the third is not: nothing at all may be printed. A program that prints as it goes has already acted on half of a bad file."},
			{Name: "eight nested lists", Gen: generated(func() (string, string) { in := tlvNest(8, tlvRec(2, "\x00\x00\x00\x2a")); return in, tlvWant(in) })},
			{Name: "nine nested lists", Gen: generated(func() (string, string) { return tlvNest(9, tlvRec(2, "\x00\x00\x00\x2a")), "" }), WantExit: 1, Stderr: true,
				Explain: "Without a depth limit, a few hundred kilobytes of nested lists make a recursive parser overflow its stack."},
			{Name: "exactly one megabyte", Gen: generated(func() (string, string) { in := tlvMegabyte(0); return in, tlvWant(in) }),
				Explain: "1,048,576 bytes is the limit itself, and is allowed."},
			{Name: "over a megabyte", Gen: generated(func() (string, string) { return tlvMegabyte(1), "" }), WantExit: 1, Stderr: true,
				Explain: "One byte over the limit must be refused, even though every record in it is valid: limits are what keep a parser's memory and time bounded."},
		},
		Hints: []string{
			"Read all of standard input with fread into a buffer of 1,048,577 bytes: if it fills completely, the input is too long.",
			"A recursive function parse(const unsigned char *p, size_t n, int depth, int print) returns 0 for a bad input; it handles a list by calling itself on the list's value with depth + 1.",
			"For the length: size_t len = (size_t)p[1] << 8 | p[2]; then if (len > n - 3) the input is bad (n is at least 3 here, so n - 3 cannot wrap).",
		},
	},
	{
		ID: "secret_scan", Stage: 17, Concept: "Finding Secrets: Patterns, Entropy & Git History",
		Title: "A secret scanner",
		Task:  "Read text from standard input and report likely secrets, one finding per line as LINE:COL KIND (both counted from 1, COL in bytes), in order of line and then column, followed by findings: N. Never print the secret itself. Rules: aws-access-key is AKIA followed by exactly 16 capital letters or digits; github-token is ghp_ followed by exactly 36 letters or digits; for both, the character just before and just after (if any) must not be a letter or digit. private-key is a line containing -----BEGIN followed by a space and, later on the same line, PRIVATE KEY-----; COL is where -----BEGIN starts. assignment: an identifier (a run of letters, digits and _) that contains, ignoring case, password, secret, token or api_key, then optionally one quote character (\" or '), optional spaces or tabs, = or : or :=, optional spaces or tabs, and a quoted value (\"…\" or '…', ending at the next same quote on the line); it is a finding when the value is at least 8 bytes long and no other finding starts inside it, and COL is the value's first byte. Scanning for assignments resumes after a quoted value's closing quote. Lines can be 100,000 bytes long.",
		File:  "secret_scan.c",
		Starter: `/* Checked exercise: secret_scan
 *   input:   aws_key = AKIAIOSFODNN7EXAMPLE
 *            db_password: "correct horse battery"
 *            name = "AKIAIOSFODNN7EXAMPLEX"
 *   output:  1:11 aws-access-key
 *            2:15 assignment
 *            findings: 2
 * Line 3 is not a key: after AKIA and its 16 characters comes another
 * letter. Nor is it an assignment: name contains none of the four
 * words.
 * Print where a secret is, never what it is: the scanner's report ends
 * up in CI logs.
 * Lines can be very long: fgets with a small buffer splits them and
 * shifts every line number after. getline(&buf, &cap, stdin) grows its
 * buffer for you.
 */
#define _POSIX_C_SOURCE 200809L
#include <ctype.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "aws_key = AKIAIOSFODNN7EXAMPLE\ndb_password: \"correct horse battery\"\nname = \"AKIAIOSFODNN7EXAMPLEX\"\n", Want: secretWant("aws_key = AKIAIOSFODNN7EXAMPLE\ndb_password: \"correct horse battery\"\nname = \"AKIAIOSFODNN7EXAMPLEX\"\n")},
			{Name: "every kind", Gen: generated(func() (string, string) {
				in := "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAA\n-----END OPENSSH PRIVATE KEY-----\n" +
					"export GITHUB=ghp_" + secretToken(36, 5) + "\n" +
					"{\"api_key\": \"9f86d081884c7d65\"}\n" +
					"SECRET_TOKEN := 'abc'  // too short to report\n"
				return in, secretWant(in)
			}),
				Explain: "A JSON key in quotes still counts (\"api_key\": \"…\"), the END line of a key is not a second finding, and a value under 8 bytes is not reported."},
			{Name: "lookalikes", Gen: generated(func() (string, string) {
				in := "XAKIAIOSFODNN7EXAMPLE\nAKIAIOSFODNN7EXAMPLE1\nAKIAIOSFODNN7EXAMPL\nakiaiosfodnn7example\nmyghp_" + secretToken(36, 6) + "\nghp_" + secretToken(35, 7) + "\n"
				return in, secretWant(in)
			}),
				Explain: "None of these is a key: a letter before AKIA, 17 characters after it, 15 after it, lower case, a letter before ghp_, and 35 characters after it. Without the boundary checks a scanner floods CI with false alarms, and people learn to ignore it."},
			{Name: "upper and mixed case names", Stdin: "DB_PASSWORD=\"s3cr3t-Pa55w0rd\"\nSlackToken: 'xoxb-1234-5678-abcdefgh'\npasswd = \"not matched, passwd is not a keyword\"\n", Want: secretWant("DB_PASSWORD=\"s3cr3t-Pa55w0rd\"\nSlackToken: 'xoxb-1234-5678-abcdefgh'\npasswd = \"not matched, passwd is not a keyword\"\n"),
				Explain: "Environment variables are usually upper case: compare the identifier ignoring case (tolower each byte)."},
			{Name: "two on one line", Gen: generated(func() (string, string) {
				in := "creds = {\"user\": \"admin\", \"password\": \"hunter2hunter2\", \"token\": \"tok_" + secretToken(20, 8) + "\"} # AKIAQWERTYUIOPASDFGH\n"
				return in, secretWant(in)
			}),
				Explain: "Keep scanning after the first finding: this line holds three secrets, and user is not one."},
			{Name: "a key in an assignment", Gen: generated(func() (string, string) {
				in := "aws_secret = \"AKIAZZZZZZZZZZZZZZZZ\"\naws_secret = \"key AKIAZZZZZZZZZZZZZZZZ\"\n"
				return in, secretWant(in)
			}),
				Explain: "When the value holds a key that has its own rule, the key is reported and the assignment is not, so one secret gives one finding."},
			{Name: "a very long line", Gen: generated(func() (string, string) {
				in := "x = 1\n" + strings.Repeat("a", 60000) + " token=\"" + secretToken(40, 9) + "\"\nAKIA" + secretToken(16, 10) + "\n"
				return in, secretWant(in)
			}),
				Explain: "A 60,000-byte line, as minified JavaScript has: a fixed buffer for fgets splits it, misses the token, and numbers every later line wrongly."},
			{Name: "nothing to report", Stdin: "func main() {\n\tfmt.Println(\"hello\")\n}\n", Want: "findings: 0\n"},
			{Name: "a large repository", Gen: generated(secretManyCase),
				Explain: "Five megabytes of code with secrets scattered through it: scanning each line once is fast enough."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"First collect a line's findings into a small array of (column, kind), then sort them by column before printing. Call qsort only when there are two or more: passing it a NULL array is undefined even with a count of 0, and the sanitizers say so.",
			"For the boundary checks, write int alnum(int c) { return isalnum((unsigned char)c); } and test the byte before the match (when it is not at column 1) and the byte after it.",
			"For assignments, walk the line identifier by identifier; after one containing a keyword, step over an optional quote, spaces, the =, : or :=, more spaces, and look for the opening quote.",
		},
	},
	{
		ID: "sha256", Stage: 17, Concept: "Integrity Checks: Hashes, Checksums & Signed Releases",
		Title: "SHA-256 and checksum verification",
		Task:  "Implement SHA-256 (FIPS 180-4). With no arguments, ./sha256 hashes standard input and prints the 64 lowercase hex digits. ./sha256 FILE prints the digest, two spaces and FILE, as sha256sum does. ./sha256 -c HEX FILE checks FILE against the expected digest HEX (64 hex digits, in either case): it prints FILE: OK when they match, and otherwise prints FILE: FAILED on standard error and exits with status 1. A file that cannot be opened gets a message on standard error and exit status 1; any other arguments, or a HEX that is not 64 hex digits, get a usage message and exit status 2. Read in pieces (files can be larger than memory), and treat the data as bytes: it can contain zeros.",
		File:  "sha256.c",
		Starter: `/* Checked exercise: sha256
 *   printf 'abc' | ./sha256
 *   ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
 * The algorithm, in FIPS 180-4 section 6.2:
 *   pad: append the byte 0x80, then zeros until the length is 56 mod 64,
 *        then the message length IN BITS as a 64-bit big-endian number
 *        (a message of 56 bytes or more needs one extra block for it)
 *   each 64-byte block: w[0..15] from the block (big-endian), then
 *        w[t] = s1(w[t-2]) + w[t-7] + s0(w[t-15]) + w[t-16]
 *        and 64 rounds updating a..h with the constants K[0..63]
 * Use uint32_t for everything: its arithmetic wraps exactly as the
 * standard says. ROTR(x, n) = (x >> n) | (x << (32 - n)).
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

const uint32_t K[64] = {
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
};

/* The initial hash value H(0). */
const uint32_t H0[8] = {
    0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
};
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "abc", Want: sha256Hex("abc") + "\n"},
			{Name: "empty input", Stdin: "", Want: sha256Hex("") + "\n",
				Explain: "Even no data is padded to one full block: 0x80, 55 zeros and a length of 0."},
			{Name: "55 bytes", Stdin: strings.Repeat("a", 55), Want: sha256Hex(strings.Repeat("a", 55)) + "\n",
				Explain: "55 bytes, the 0x80 byte and the 8-byte length fill exactly one block."},
			{Name: "56 bytes", Stdin: strings.Repeat("a", 56), Want: sha256Hex(strings.Repeat("a", 56)) + "\n",
				Explain: "At 56 bytes the length no longer fits after the 0x80 byte: the padding needs a second block. Off-by-one padding fails here and only here, so test the boundaries."},
			{Name: "64 bytes", Stdin: strings.Repeat("b", 64), Want: sha256Hex(strings.Repeat("b", 64)) + "\n"},
			{Name: "a file", Args: []string{"release.tar"}, Files: map[string]string{"release.tar": sha256Data(100000, 21)}, Want: sha256Hex(sha256Data(100000, 21)) + "  release.tar\n",
				Explain: "Binary data contains zero bytes: reading it with string functions (fgets, strlen) stops at the first one."},
			{Name: "a million a's", Gen: generated(func() (string, string) {
				in := strings.Repeat("a", 1000000)
				return in, sha256Hex(in) + "\n"
			}),
				Explain: "The test vector from the standard: cdc76e5c…. It needs the length in bits (8,000,000), not bytes, and many blocks read in pieces."},
			{Name: "a matching checksum", Args: []string{"-c", strings.ToUpper(sha256Hex(sha256Data(5000, 22))), "app.bin"}, Files: map[string]string{"app.bin": sha256Data(5000, 22)}, Want: "app.bin: OK\n",
				Explain: "Published checksums come in both cases: compare the hex digits ignoring case."},
			{Name: "a tampered file", Args: []string{"-c", sha256Hex(sha256Data(5000, 22)), "app.bin"}, Files: map[string]string{"app.bin": sha256Data(5000, 22)[:4999] + "!"}, WantExit: 1, Stderr: true,
				Explain: "One changed byte gives a completely different digest: report FAILED, and exit with a non-zero status so a script that installs the file stops."},
			{Name: "a malformed checksum", Args: []string{"-c", "1234", "app.bin"}, Files: map[string]string{"app.bin": "x"}, WantExit: 2, Stderr: true,
				Explain: "A checksum that is not 64 hex digits is a mistake by whoever typed it, not a mismatch: refuse it as bad usage."},
			{Name: "a missing file", Args: []string{"nothere.bin"}, WantExit: 1, Stderr: true},
			{Name: "16 megabytes", Gen: generated(func() (string, string) { in := sha256Data(16<<20, 23); return in, sha256Hex(in) + "\n" }),
				Explain: "A disk image's worth of data, read in pieces: loading it all with one malloc works here but not for a file larger than memory."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Keep a state struct: the eight hash words, a 64-byte block buffer with a count of bytes in it, and the total length as a uint64_t. update() fills the buffer and compresses each full block; final() pads.",
			"Read with fread into a 64 KB buffer in a loop: it returns how many bytes it read, zeros included.",
			"Print the digest with printf(\"%08x\", h[i]) for i from 0 to 7.",
		},
	},
}

// tlvRec is one record: a type, a big-endian two-byte length, the value.
func tlvRec(t byte, v string) string {
	return string([]byte{t, byte(len(v) >> 8), byte(len(v))}) + v
}

// tlvNest wraps inner in depth lists.
func tlvNest(depth int, inner string) string {
	for i := 0; i < depth; i++ {
		inner = tlvRec(3, inner)
	}
	return inner
}

// tlvWant parses as the task describes; an invalid input gives "".
func tlvWant(in string) string {
	if len(in) > 1<<20 {
		return ""
	}
	var b strings.Builder
	var parse func(v string, depth int) bool
	parse = func(v string, depth int) bool {
		for len(v) > 0 {
			if len(v) < 3 {
				return false
			}
			t, n := v[0], int(v[1])<<8|int(v[2])
			v = v[3:]
			if n > len(v) {
				return false
			}
			val := v[:n]
			v = v[n:]
			ind := strings.Repeat("  ", depth)
			switch t {
			case 1:
				b.WriteString(ind + "text \"")
				for i := 0; i < len(val); i++ {
					c := val[i]
					if c >= ' ' && c <= '~' && c != '"' && c != '\\' {
						b.WriteByte(c)
					} else {
						fmt.Fprintf(&b, "\\x%02x", c)
					}
				}
				b.WriteString("\"\n")
			case 2:
				if n != 4 {
					return false
				}
				fmt.Fprintf(&b, "%snumber %d\n", ind, uint32(val[0])<<24|uint32(val[1])<<16|uint32(val[2])<<8|uint32(val[3]))
			case 3:
				if depth >= 8 {
					return false
				}
				b.WriteString(ind + "list {\n")
				if !parse(val, depth+1) {
					return false
				}
				b.WriteString(ind + "}\n")
			default:
				fmt.Fprintf(&b, "%sskip type %d, %d bytes\n", ind, t, n)
			}
		}
		return true
	}
	if !parse(in, 0) {
		return ""
	}
	return b.String()
}

// tlvMegabyte is 1,048,576 + extra bytes of valid records, mostly
// skipped ones, so the output stays small.
func tlvMegabyte(extra int) string {
	rng := rand.New(rand.NewSource(31))
	var b strings.Builder
	const limit = 1 << 20
	for b.Len() < limit-60000 {
		switch rng.Intn(4) {
		case 0:
			b.WriteString(tlvRec(1, fmt.Sprintf("record %d", b.Len())))
		case 1:
			v := rng.Uint32()
			b.WriteString(tlvRec(2, string([]byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)})))
		case 2:
			b.WriteString(tlvRec(3, tlvRec(1, "inner")+tlvRec(byte(10+rng.Intn(200)), strings.Repeat("\x00", rng.Intn(300)))))
		default:
			b.WriteString(tlvRec(byte(4+rng.Intn(250)), strings.Repeat("z", 200+rng.Intn(2000))))
		}
	}
	// Pad to the exact limit with one last record to skip.
	b.WriteString(tlvRec(200, strings.Repeat("p", limit+extra-b.Len()-3)))
	return b.String()
}

// secretToken is n letters and digits, deterministic per seed.
func secretToken(n int, seed int64) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	rng := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rng.Intn(len(chars))]
	}
	// Capital letters and digits only for AWS-style keys of 16.
	if n == 16 {
		for i := range b {
			b[i] = chars[rng.Intn(36)]
		}
	}
	return string(b)
}

type secretFinding struct {
	col  int
	kind string
}

func secretAlnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

func secretIdent(c byte) bool { return secretAlnum(c) || c == '_' }

// secretPrefixed finds prefix followed by exactly n bytes for which ok
// holds, with no letter or digit on either side.
func secretPrefixed(line, prefix string, n int, ok func(byte) bool, kind string) []secretFinding {
	var out []secretFinding
	for i := 0; i+len(prefix)+n <= len(line); i++ {
		if line[i:i+len(prefix)] != prefix || i > 0 && secretAlnum(line[i-1]) {
			continue
		}
		end := i + len(prefix) + n
		good := end == len(line) || !secretAlnum(line[end])
		for k := i + len(prefix); good && k < end; k++ {
			good = ok(line[k])
		}
		if good {
			out = append(out, secretFinding{i + 1, kind})
		}
	}
	return out
}

// secretLine returns one line's findings, sorted by column.
func secretLine(line string) []secretFinding {
	upperDigit := func(c byte) bool { return c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' }
	found := secretPrefixed(line, "AKIA", 16, upperDigit, "aws-access-key")
	found = append(found, secretPrefixed(line, "ghp_", 36, secretAlnum, "github-token")...)
	if i := strings.Index(line, "-----BEGIN "); i >= 0 && strings.Contains(line[i:], "PRIVATE KEY-----") {
		found = append(found, secretFinding{i + 1, "private-key"})
	}
	specific := append([]secretFinding(nil), found...)
	for i := 0; i < len(line); {
		if !secretIdent(line[i]) || i > 0 && secretIdent(line[i-1]) {
			i++
			continue
		}
		j := i
		for j < len(line) && secretIdent(line[j]) {
			j++
		}
		word := strings.ToLower(line[i:j])
		next := j
		if strings.Contains(word, "password") || strings.Contains(word, "secret") || strings.Contains(word, "token") || strings.Contains(word, "api_key") {
			k := j
			if k < len(line) && (line[k] == '"' || line[k] == '\'') {
				k++
			}
			for k < len(line) && (line[k] == ' ' || line[k] == '\t') {
				k++
			}
			sep := false
			if k+1 < len(line) && line[k] == ':' && line[k+1] == '=' {
				k, sep = k+2, true
			} else if k < len(line) && (line[k] == '=' || line[k] == ':') {
				k, sep = k+1, true
			}
			for sep && k < len(line) && (line[k] == ' ' || line[k] == '\t') {
				k++
			}
			if sep && k < len(line) && (line[k] == '"' || line[k] == '\'') {
				if end := strings.IndexByte(line[k+1:], line[k]); end >= 0 {
					start, stop := k+1, k+1+end // the value is line[start:stop]
					inside := false
					for _, f := range specific {
						inside = inside || f.col-1 >= start && f.col-1 < stop
					}
					if stop-start >= 8 && !inside {
						found = append(found, secretFinding{start + 1, "assignment"})
					}
					next = stop + 1
				}
			}
		}
		i = next
	}
	sort.SliceStable(found, func(a, b int) bool { return found[a].col < found[b].col })
	return found
}

// secretWant scans text as the task describes.
func secretWant(in string) string {
	in = strings.TrimSuffix(in, "\n")
	var b strings.Builder
	n := 0
	if in != "" {
		for i, line := range strings.Split(in, "\n") {
			for _, f := range secretLine(line) {
				fmt.Fprintf(&b, "%d:%d %s\n", i+1, f.col, f.kind)
				n++
			}
		}
	}
	fmt.Fprintf(&b, "findings: %d\n", n)
	return b.String()
}

// secretManyCase is about 5 MB of code-like text with secrets planted
// on a few hundred lines, and lookalikes on others.
func secretManyCase() (string, string) {
	rng := rand.New(rand.NewSource(41))
	words := []string{"func", "return", "token", "password", "value", "config", "if", "for", "secretary", "user", "AKIA", "ghp_"}
	var b strings.Builder
	for line := 0; b.Len() < 5<<20; line++ {
		switch r := rng.Intn(1000); {
		case r < 3:
			fmt.Fprintf(&b, "aws_access_key_id = AKIA%s\n", secretToken(16, int64(line)))
		case r < 5:
			fmt.Fprintf(&b, "\t%sToken := \"%s\"\n", words[rng.Intn(len(words))], secretToken(8+rng.Intn(30), int64(line)))
		case r < 6:
			fmt.Fprintf(&b, "GH=ghp_%s # ci\n", secretToken(36, int64(line)))
		case r < 8:
			fmt.Fprintf(&b, "x := \"AKIA%s\" // not a key\n", secretToken(17, int64(line)))
		default:
			n := 3 + rng.Intn(12)
			for k := 0; k < n; k++ {
				if k > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(words[rng.Intn(len(words))])
				if rng.Intn(6) == 0 {
					fmt.Fprintf(&b, " = \"%s\"", secretToken(rng.Intn(12), int64(line*31+k)))
				}
			}
			b.WriteByte('\n')
		}
	}
	in := b.String()
	return in, secretWant(in)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// sha256Data is n bytes of deterministic binary data, zeros included.
func sha256Data(n int, seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	rng.Read(b)
	for i := 0; i < n; i += 97 {
		b[i] = 0
	}
	return string(b)
}
