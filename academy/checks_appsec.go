package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Checked exercises for Stage 10 (security): storing passwords with a
// slow, salted hash, and encoding output so data never becomes code.
var appsecChecks = []Check{
	{
		ID: "pbkdf2", Stage: 10, Concept: "Hashing & Password Storage",
		Title: "Password hashing with PBKDF2",
		Task:  "Implement PBKDF2-HMAC-SHA256 (RFC 8018) with your own SHA-256 and HMAC, deriving 32 bytes. ./pbkdf2 hash ITERATIONS SALTHEX reads a password (the first line of standard input, without its newline) and prints pbkdf2-sha256$ITERATIONS$SALTHEX$HASHHEX, the hash in lowercase hex. ./pbkdf2 verify reads a stored string in that form on the first line and a password on the second, recomputes the hash with the stored iterations and salt, and prints ok if it matches; otherwise it prints a message on standard error and exits with status 1. ITERATIONS is 1 to 1,000,000; the salt is 0 to 64 bytes of lowercase hex; passwords can be any bytes except a newline, and longer than 64 bytes. A malformed command, iteration count, salt or stored string gets a message on standard error and exit status 2.",
		File:  "pbkdf2.c",
		Starter: `/* Checked exercise: pbkdf2
 *   printf 'passwd\n' | ./pbkdf2 hash 1 73616c74        (73616c74 is "salt")
 *   pbkdf2-sha256$1$73616c74$55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc
 * For 32 bytes of output, PBKDF2 is one block:
 *   U1 = HMAC(password, salt || 00 00 00 01)
 *   U2 = HMAC(password, U1), ..., Uc = HMAC(password, Uc-1)
 *   result = U1 ^ U2 ^ ... ^ Uc          (c = ITERATIONS)
 * HMAC(K, m) = SHA256((K' ^ 0x5c x 64) || SHA256((K' ^ 0x36 x 64) || m)),
 * with K' the key padded to 64 bytes, or SHA256(K) padded when K is
 * longer. Your sha256.c from Stage 17 has the hash; with hundreds of
 * thousands of iterations, avoid malloc inside the loop.
 * In verify, use the iteration count stored with the hash: that is how
 * old hashes keep working after you raise the count for new ones. And
 * compare the final bytes in constant time.
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"hash", "1", "73616c74"}, Stdin: "passwd\n", Want: pbkdf2Stored("passwd", "salt", 1) + "\n"},
			{Name: "the standard's test vector", Args: []string{"hash", "80000", hex.EncodeToString([]byte("NaCl"))}, Stdin: "Password\n", Want: pbkdf2Stored("Password", "NaCl", 80000) + "\n",
				Explain: "RFC 7914's second PBKDF2-HMAC-SHA256 vector, 80,000 iterations, cut to 32 bytes: 4ddcd8f60b98be21…."},
			{Name: "a password longer than a block", Args: []string{"hash", "1000", "0011223344556677"}, Stdin: strings.Repeat("long passphrase ", 6) + "\n", Want: pbkdf2Stored(strings.Repeat("long passphrase ", 6), "\x00\x11\x22\x33\x44\x55\x66\x77", 1000) + "\n",
				Explain: "96 bytes is more than SHA-256's 64-byte block: HMAC hashes such a key first. Cutting or padding it gives a different hash, and people do use long passphrases."},
			{Name: "an empty salt", Args: []string{"hash", "10", ""}, Stdin: "x\n", Want: pbkdf2Stored("x", "", 10) + "\n"},
			{Name: "the right password", Args: []string{"verify"}, Stdin: pbkdf2Stored("correct horse", "\xa5\x1b\x07\x99\x10\xfe\x33\x42", 20000) + "\ncorrect horse\n", Want: "ok\n"},
			{Name: "the wrong password", Args: []string{"verify"}, Stdin: pbkdf2Stored("correct horse", "\xa5\x1b\x07\x99\x10\xfe\x33\x42", 20000) + "\ncorrect horsf\n", WantExit: 1, Stderr: true},
			{Name: "the stored iteration count", Args: []string{"verify"}, Stdin: pbkdf2Stored("legacy", "\x01\x02", 1000) + "\nlegacy\n", Want: "ok\n",
				Explain: "This hash was stored with 1,000 iterations, from before the count was raised. Verifying with the count you would use today gives a different answer and locks every older account out."},
			{Name: "spaces in a password", Args: []string{"verify"}, Stdin: pbkdf2Stored("  two  spaces ", "\x05", 50) + "\n  two  spaces \n", Want: "ok\n",
				Explain: "Spaces at either end are part of the password: remove only the newline."},
			{Name: "a malformed stored string", Args: []string{"verify"}, Stdin: "pbkdf2-sha256$abc$00$00\nx\n", WantExit: 2, Stderr: true},
			{Name: "too many iterations", Args: []string{"hash", "2000000", "00"}, Stdin: "x\n", WantExit: 2, Stderr: true,
				Explain: "A cap on the iteration count keeps a stored string from making verification take minutes."},
			{Name: "two hundred thousand iterations", Args: []string{"hash", "200000", "6c6f6e67"}, Stdin: "patience\n", Want: pbkdf2Stored("patience", "long", 200000) + "\n",
				Explain: "Hashing the key into the inner and outer pads once, and only the 32-byte U values in the loop, keeps this to about a second even with the sanitizers on."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Write hmac_sha256(key, klen, msg, mlen, out) first and test it against a known value, then the PBKDF2 loop over it.",
			"For speed, hash the inner and outer padded keys once, and keep the two partial SHA-256 states; each iteration then hashes only 32 bytes into each.",
			"Constant-time compare: unsigned diff = 0; for each byte diff |= a[i] ^ b[i]; then test diff once.",
		},
	},
	{
		ID: "html_escape", Stage: 10, Concept: "Web Vulnerabilities: Injection & XSS",
		Title: "Output encoding for HTML and URLs",
		Task:  "./html_escape html copies standard input to standard output, replacing & with &amp;, < with &lt;, > with &gt;, \" with &quot; and ' with &#39;, so text is safe both between tags and inside a quoted attribute. ./html_escape url percent-encodes every byte except the unreserved characters A-Z a-z 0-9 - . _ ~ (RFC 3986), writing %HH with two upper-case hex digits; a space becomes %20. Every other byte, including newlines, zeros and bytes of UTF-8 above 127, is copied (html) or encoded (url) as it is. Any other argument gets a usage message on standard error and exit status 2.",
		File:  "html_escape.c",
		Starter: `/* Checked exercise: html_escape
 *   printf '<b>"Tom & Jerry"</b>' | ./html_escape html
 *   &lt;b&gt;&quot;Tom &amp; Jerry&quot;&lt;/b&gt;
 *   printf 'a b/é' | ./html_escape url
 *   a%20b%2F%C3%A9
 * Read with getchar into an int, so that bytes above 127 arrive as 128
 * to 255 and EOF stays apart from them. Printing a char of 0xC3 with
 * %02X can give FFFFFFC3 where char is signed: use unsigned values.
 * Encode where the data goes, at the last moment: HTML escaping for a
 * page, percent-encoding for a URL. Neither replaces the other.
 */
#include <stdio.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"html"}, Stdin: "<b>\"Tom & Jerry\"</b>", Want: "&lt;b&gt;&quot;Tom &amp; Jerry&quot;&lt;/b&gt;"},
			{Name: "an attribute", Args: []string{"html"}, Stdin: "x' onmouseover='alert(1)", Want: htmlEscapeWant("x' onmouseover='alert(1)"),
				Explain: "Inside value='…', an unescaped single quote ends the attribute, and the rest becomes a new attribute: an event handler. Escape ' as &#39; as well as \"."},
			{Name: "already escaped", Args: []string{"html"}, Stdin: "&lt;b&gt; &amp;amp;", Want: htmlEscapeWant("&lt;b&gt; &amp;amp;"),
				Explain: "Escaping is not idempotent: &lt; becomes &amp;lt;, so the page shows &lt; as the user typed it. Escape once, at output, never twice."},
			{Name: "every byte", Args: []string{"html"}, Gen: generated(func() (string, string) {
				var b strings.Builder
				for c := 0; c < 256; c++ {
					b.WriteByte(byte(c))
				}
				return b.String(), htmlEscapeWant(b.String())
			}),
				Explain: "All 256 byte values, zero included: only the five special characters change, and reading stops only at end of file."},
			{Name: "a URL value", Args: []string{"url"}, Stdin: "a b/é?x=1&y=2#top", Want: urlEncodeWant("a b/é?x=1&y=2#top"),
				Explain: "/, ?, & and # have meanings in a URL, so inside a value they must be encoded, or the value ends early. é is two UTF-8 bytes, encoded one by one: %C3%A9."},
			{Name: "unreserved characters", Args: []string{"url"}, Stdin: "AZaz09-._~", Want: "AZaz09-._~"},
			{Name: "high bytes", Args: []string{"url"}, Stdin: "\xff\x80\x00\n", Want: "%FF%80%00%0A",
				Explain: "With a signed char, 0xFF is -1 and prints as %FFFFFFFF, or stops the loop if compared with EOF: keep bytes as unsigned values in an int."},
			{Name: "200 KB of text", Args: []string{"html"}, Gen: generated(func() (string, string) {
				rng := rand.New(rand.NewSource(131))
				var b strings.Builder
				for b.Len() < 200000 {
					b.WriteString([]string{"plain words ", "<tag> ", "a & b ", "\"quoted\" ", "it's ", "\n"}[rng.Intn(6)])
				}
				return b.String(), htmlEscapeWant(b.String())
			})},
			{Name: "an unknown mode", Args: []string{"js"}, WantExit: 2, Stderr: true,
				Explain: "Encoding for JavaScript strings has its own rules; refusing an unknown context is safer than guessing."},
		},
		Hints: []string{
			"int c; while ((c = getchar()) != EOF) { switch (c) { case '&': fputs(\"&amp;\", stdout); break; ... default: putchar(c); } }",
			"For url mode: if c is unreserved, putchar(c); otherwise printf(\"%%%02X\", c) with c an int from 0 to 255.",
			"Check argc and the mode before reading anything.",
		},
	},
}

// pbkdf2Key derives 32 bytes with PBKDF2-HMAC-SHA256 (one block).
func pbkdf2Key(password, salt string, iter int) []byte {
	mac := hmac.New(sha256.New, []byte(password))
	var counter [4]byte
	binary.BigEndian.PutUint32(counter[:], 1)
	mac.Write([]byte(salt))
	mac.Write(counter[:])
	u := mac.Sum(nil)
	out := append([]byte(nil), u...)
	for i := 1; i < iter; i++ {
		mac.Reset()
		mac.Write(u)
		u = mac.Sum(nil)
		for k := range out {
			out[k] ^= u[k]
		}
	}
	return out
}

func pbkdf2Stored(password, salt string, iter int) string {
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", iter, hex.EncodeToString([]byte(salt)), hex.EncodeToString(pbkdf2Key(password, salt, iter)))
}

func htmlEscapeWant(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&#39;").Replace(s)
}

func urlEncodeWant(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || strings.IndexByte("-._~", c) >= 0 {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}
