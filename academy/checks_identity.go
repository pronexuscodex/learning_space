package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Checked exercises for Stage 18 (identity and cloud security): TOTP
// codes, JWT verification and an IAM-style policy evaluator, in C.
var identityChecks = []Check{
	{
		ID: "totp", Stage: 18, Concept: "Strong Authentication: MFA & Passkeys",
		Title: "TOTP codes, from SHA-1 up",
		Task:  "Implement TOTP (RFC 6238) with your own SHA-1 and HMAC. Each line of standard input is SECRET TIME, or SECRET TIME CODE. SECRET is Base32 (A to Z and 2 to 7, in either case, optionally ending in = padding) of at least 16 characters; TIME is a whole number of Unix seconds from 0 to 1,000,000,000,000; CODE is exactly 6 digits. The step is TIME / 30; the code is HMAC-SHA1 of the step as 8 big-endian bytes, keyed with the decoded secret, truncated as RFC 4226 says (4 bytes at the offset given by the low 4 bits of the last byte, top bit cleared) and taken mod 1,000,000. For SECRET TIME, print the code as 6 digits. For SECRET TIME CODE, check the code against the current step, then the one before, then the one after, and print ok 0, ok -1 or ok +1 for the first that matches, or no. Blank lines are skipped. A line that breaks these rules gets a message on standard error and exit status 1.",
		File:  "totp.c",
		Starter: `/* Checked exercise: totp
 *   input:  GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59
 *           GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59 287082
 *   output: 287082
 *           ok 0
 * The secret is the RFC's test key, "12345678901234567890" in Base32.
 * 1. Base32: each character is 5 bits (A = 0 ... Z = 25, 2 = 26 ... 7 = 31);
 *    collect bits and emit a byte whenever 8 are waiting.
 * 2. SHA-1 (FIPS 180-4): like SHA-256 but five words, 80 rounds, and
 *    w[t] = ROTL1(w[t-3] ^ w[t-8] ^ w[t-14] ^ w[t-16]).
 * 3. HMAC(K, m) = SHA1((K' ^ opad) || SHA1((K' ^ ipad) || m)), where K'
 *    is K padded with zeros to 64 bytes, or SHA1(K) padded, if K is longer.
 * 4. step = TIME / 30, as 8 bytes, most significant first: a 64-bit value.
 * 5. o = mac[19] & 15; n = (mac[o] & 0x7f) << 24 | mac[o+1] << 16 | ...;
 *    code = n % 1000000, printed with %06u (codes can start with 0).
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59 287082\n", Want: totpWant("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59 287082\n")},
			{Name: "the standard's test vectors", Stdin: totpVectors, Want: totpWant(totpVectors),
				Explain: "The six times from RFC 6238's table, with the SHA-1 key. 1234567890 gives 005924: print codes with %06u, or the leading zeros vanish and the code no longer matches what the app shows."},
			{Name: "after 2038", Stdin: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 2200000000\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 200000000000\n", Want: totpWant("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 2200000000\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 200000000000\n"),
				Explain: "2,200,000,000 seconds does not fit in a signed 32-bit int (atoi overflows), and 200,000,000,000 seconds is a step above 2³²: parse the time into a 64-bit integer and send all 8 bytes of the step."},
			{Name: "the top bit", Gen: generated(totpTopBitCase),
				Explain: "These steps' truncated bytes start at or above 0x80. Without clearing the top bit, the 31-bit number becomes negative or too large, and the code is wrong."},
			{Name: "the time window", Gen: generated(totpWindowCase),
				Explain: "A code from the previous or the next step is accepted (clocks drift and people type slowly), one from two steps away is not, and the current step is checked first."},
			{Name: "lower case and padding", Stdin: "gezdgnbvgy3tqojqgezdgnbvgy3tqojq 1111111109\nJBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP======== 1700000000\n", Want: totpWant("gezdgnbvgy3tqojqgezdgnbvgy3tqojq 1111111109\nJBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP======== 1700000000\n")},
			{Name: "a long secret", Gen: generated(func() (string, string) {
				in := totpSecret(128, 3) + " 1700000000\n"
				return in, totpWant(in)
			}),
				Explain: "128 Base32 characters are 80 bytes, more than SHA-1's 64-byte block: HMAC hashes such a key first. Padding or cutting it instead gives a different code."},
			{Name: "not Base32", Stdin: "GEZDGNBVGY3TQOJ1GEZDGNBVGY3TQOJQ 59\n", WantExit: 1, Stderr: true,
				Explain: "1, 0, 8 and 9 are not Base32 (they look like I, O, B and g). A secret with one is a typing mistake: refuse it rather than decode garbage."},
			{Name: "a secret that is too short", Stdin: "GEZDGNBVGY3T 59\n", WantExit: 1, Stderr: true,
				Explain: "12 characters are 60 bits: too few to resist guessing the key itself. RFC 4226 asks for at least 128 bits, and 160 is usual."},
			{Name: "a code of 5 digits", Stdin: "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59 28708\n", WantExit: 1, Stderr: true},
			{Name: "fifty thousand codes", Gen: generated(totpManyCase),
				Explain: "Four SHA-1 blocks per code: well under a second in C."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read each line with fgets into a 512-byte buffer and split it with sscanf(line, \"%300s %30s %30s\", ...), checking how many fields matched.",
			"Parse TIME with strtoull and check that the whole field was digits and at most 1,000,000,000,000.",
			"Write sha1(const unsigned char *data, size_t n, unsigned char out[20]) for messages that fit in memory: HMAC's messages here are at most 64 + 20 bytes.",
		},
	},
	{
		ID: "jwt_verify", Stage: 18, Concept: "Sessions, Tokens & Single Sign-On",
		Title: "A JWT verifier that decides for itself",
		Task:  "./jwt_verify SECRET ISSUER AUDIENCE NOW reads one token, an HS256 JSON Web Token, from standard input and checks it. The token is three parts separated by dots, each base64url without padding (letters, digits, - and _). The first two decode to compact JSON objects ({\"key\":value,…} with no spaces) whose values are strings without escapes or whole numbers. The token is valid only if: the header's alg is exactly HS256; the third part decodes to the 32-byte HMAC-SHA256, keyed with SECRET, of the first two parts as they appear, joined by a dot; the payload has sub, iss and aud strings and an exp number; iss and aud equal ISSUER and AUDIENCE; NOW is less than exp, and not less than nbf if there is one; and no key appears twice in either object. Then print valid sub=SUB exp=EXP. Otherwise print invalid and the reason on standard error and exit with status 1. Wrong arguments get a usage message and exit status 2.",
		File:  "jwt_verify.c",
		Starter: `/* Checked exercise: jwt_verify
 *   ./jwt_verify my-secret https://login.example notes-app 1800000000 < token
 *   valid sub=user-4182 exp=1800000900
 * Decide the algorithm yourself: the header must say HS256, and you
 * check an HS256 signature, whatever else the header claims. A header
 * saying "none" with an empty signature is the classic forgery.
 * HMAC-SHA256(K, m) = SHA256((K' ^ 0x5c...) || SHA256((K' ^ 0x36...) || m)),
 * K' = K padded with zeros to 64 bytes, or SHA256(K) padded if K is
 * longer. Your sha256.c from Stage 17 has the hash.
 * base64url: A-Z a-z 0-9 - _ for 0..63; no '=' at the end.
 * Compare the signature in constant time: OR together the XOR of every
 * byte pair, then test the result once.
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: jwtArgs(jwtNow), Stdin: jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "\n", Want: "valid sub=user-4182 exp=1800000900\n"},
			{Name: "alg none", Args: jwtArgs(jwtNow), Stdin: jwtPart(`{"alg":"none","typ":"JWT"}`) + "." + jwtPart(jwtClaims("admin", jwtExp, "")) + ".\n", WantExit: 1, Stderr: true,
				Explain: "The forged token says alg none and carries no signature. A verifier that lets the token choose its own algorithm accepts it: decide the algorithm in your code."},
			{Name: "another algorithm", Args: jwtArgs(jwtNow), Stdin: jwtSign(`{"alg":"HS512","typ":"JWT"}`, jwtClaims("user-4182", jwtExp, "")) + "\n", WantExit: 1, Stderr: true,
				Explain: "Signed correctly with HS256, but the header claims HS512. Only the algorithm you expect may be accepted, so that a header can never switch how the signature is checked."},
			{Name: "a changed payload", Args: jwtArgs(jwtNow), Gen: generated(func() (string, string) {
				good := strings.Split(jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")), ".")
				return good[0] + "." + jwtPart(jwtClaims("admin", jwtExp, "")) + "." + good[2] + "\n", ""
			}), WantExit: 1, Stderr: true},
			{Name: "the wrong audience", Args: []string{jwtSecret, jwtIssuer, "billing-app", fmt.Sprint(jwtNow)}, Stdin: jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "\n", WantExit: 1, Stderr: true,
				Explain: "A valid token issued for notes-app, presented to billing-app. Without the audience check, any token from the same login service works on every app it serves."},
			{Name: "the wrong issuer", Args: []string{jwtSecret, "https://login.other", jwtAudience, fmt.Sprint(jwtNow)}, Stdin: jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "\n", WantExit: 1, Stderr: true},
			{Name: "expiry, to the second", Args: jwtArgs(jwtExp - 1), Stdin: jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "\n", Want: "valid sub=user-4182 exp=1800000900\n"},
			{Name: "expired", Args: jwtArgs(jwtExp), Stdin: jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "\n", WantExit: 1, Stderr: true,
				Explain: "At NOW = exp the token has expired: exp is the first moment it is no longer valid."},
			{Name: "not yet valid", Args: jwtArgs(jwtNow), Stdin: jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, `,"nbf":1800000300`)) + "\n", WantExit: 1, Stderr: true},
			{Name: "every base64url character", Args: jwtArgs(jwtNow), Gen: generated(jwtURLCharsCase),
				Explain: "This token's signature contains - and _, which standard base64 writes as + and /. Decoding with the standard alphabet rejects a genuine token."},
			{Name: "a long key", Args: []string{strings.Repeat("k3y-", 25), jwtIssuer, jwtAudience, fmt.Sprint(jwtNow)}, Stdin: jwtSignWith(strings.Repeat("k3y-", 25), jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "\n", Want: "valid sub=user-4182 exp=1800000900\n",
				Explain: "A 100-byte key is longer than SHA-256's 64-byte block: HMAC hashes it first. Truncating or padding it gives a different signature."},
			{Name: "a key given twice", Args: jwtArgs(jwtNow), Stdin: jwtSign(jwtHeader, `{"sub":"user-4182","iss":"https://login.example","aud":"other-app","exp":1800000900,"aud":"notes-app"}`) + "\n", WantExit: 1, Stderr: true,
				Explain: "Two aud values: one parser might read the first and another the last, so two parts of a system disagree about the same token. Refuse duplicates."},
			{Name: "padding", Args: jwtArgs(jwtNow), Gen: generated(func() (string, string) {
				return jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + "=\n", ""
			}), WantExit: 1, Stderr: true},
			{Name: "four parts", Args: jwtArgs(jwtNow), Gen: generated(func() (string, string) {
				return jwtSign(jwtHeader, jwtClaims("user-4182", jwtExp, "")) + ".e30\n", ""
			}), WantExit: 1, Stderr: true},
			{Name: "no expiry", Args: jwtArgs(jwtNow), Stdin: jwtSign(jwtHeader, `{"sub":"user-4182","iss":"https://login.example","aud":"notes-app"}`) + "\n", WantExit: 1, Stderr: true,
				Explain: "A token without exp would be valid forever: refuse it."},
			{Name: "no arguments", WantExit: 2, Stderr: true},
		},
		Hints: []string{
			"Split the line at the two dots first, and refuse anything else; keep pointers to the parts as they were sent, because the signature covers that text, not your decoded JSON.",
			"A tiny JSON reader is enough: expect {, then repeatedly \"key\": followed by \"string\" or digits, separated by commas, then }. Store up to 16 key/value pairs and refuse a key seen twice.",
			"Check everything, and decide at the end; the order of the checks does not change the answer.",
		},
	},
	{
		ID: "policy_eval", Stage: 18, Concept: "Policy as Code: Writing & Testing Authorization Rules",
		Title: "An IAM-style policy evaluator",
		Task:  "./policy_eval POLICY reads rules from the file POLICY, one per line: allow or deny, then a principal, an action and a resource pattern, separated by spaces or tabs. Blank lines and lines whose first non-blank character is # are ignored, but still counted as lines. In a pattern, * matches any run of characters, including none; every other character matches itself, case included. Then each line of standard input is a request, PRINCIPAL ACTION RESOURCE. For each, print deny (line N) for the first deny rule, in file order, whose three patterns all match; otherwise allow (line N) for the first allow rule that matches; otherwise deny (no rule). A rule line that is not allow or deny with exactly three patterns, or a request without exactly three fields, gets a message on standard error and exit status 1. Policies can have 10,000 rules, and patterns come from people you do not fully trust: matching must not take exponential time.",
		File:  "policy_eval.c",
		Starter: `// Checked exercise: policy_eval
// (Line comments here: the patterns below contain slash-star, which
// would start a comment inside a /* ... */ block.)
//   policy:   # who may do what
//             allow user:* storage:Get* reports/*
//             deny  user:intern storage:* reports/secret/*
//   requests: user:ana storage:GetObject reports/q1.pdf
//             user:intern storage:GetObject reports/secret/plan.pdf
//             user:ana storage:PutObject reports/q1.pdf
//   output:   allow (line 2)
//             deny (line 3)
//             deny (no rule)
// Deny always wins; with no matching rule, the answer is no.
// Matching *: walk pattern and text together; at a *, remember where
// the star is and where the text was. On a mismatch, go back to just
// after the last star and let it swallow one more character. Trying
// every possibility recursively instead takes exponential time on
// patterns such as *a*a*a*a*a*a*b.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": policyStarter},
				Stdin: "user:ana storage:GetObject reports/q1.pdf\nuser:intern storage:GetObject reports/secret/plan.pdf\nuser:ana storage:PutObject reports/q1.pdf\n",
				Want:  policyWant(policyStarter, "user:ana storage:GetObject reports/q1.pdf\nuser:intern storage:GetObject reports/secret/plan.pdf\nuser:ana storage:PutObject reports/q1.pdf\n")},
			{Name: "deny wins, wherever it is", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "allow * * *\nallow admin:* iam:* *\ndeny * iam:DeleteUser user/root\n"},
				Stdin:   "admin:bo iam:DeleteUser user/root\nadmin:bo iam:DeleteUser user/ana\nuser:cy iam:CreateUser user/x\n",
				Want:    policyWant("allow * * *\nallow admin:* iam:* *\ndeny * iam:DeleteUser user/root\n", "admin:bo iam:DeleteUser user/root\nadmin:bo iam:DeleteUser user/ana\nuser:cy iam:CreateUser user/x\n"),
				Explain: "Two allow rules match and the deny comes last in the file: deny still wins. Taking the first rule that matches, whatever it says, lets an early broad allow override every deny below it."},
			{Name: "wildcards", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": policyWildcards},
				Stdin: policyWildcardRequests, Want: policyWant(policyWildcards, policyWildcardRequests),
				Explain: "* also matches nothing (logs/* matches logs/), several stars work together, and a star must be able to give back characters: *ab matches aab only if the star takes one a, not two."},
			{Name: "comments count as lines", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "# rules for the notes app\n\n   # admins\nallow role:admin notes:* *\n\t\ndeny role:guest notes:Write* *\n"},
				Stdin:   "role:admin notes:Delete n/1\nrole:guest notes:WriteNote n/2\n",
				Want:    policyWant("# rules for the notes app\n\n   # admins\nallow role:admin notes:* *\n\t\ndeny role:guest notes:Write* *\n", "role:admin notes:Delete n/1\nrole:guest notes:WriteNote n/2\n"),
				Explain: "Line numbers are the file's own, so a reviewer can find the rule: count every line, comments and blank ones included."},
			{Name: "case matters", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "allow user:* storage:Get* *\n"},
				Stdin: "user:ana storage:getobject x\nUser:ana storage:GetObject x\n", Want: "deny (no rule)\ndeny (no rule)\n"},
			{Name: "a pathological pattern", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "allow * *a*a*a*a*a*a*b *\nallow * * *a*a*a*a*a*a*b\n"},
				Gen: generated(func() (string, string) {
					var b strings.Builder
					for i := 0; i < 5; i++ {
						fmt.Fprintf(&b, "p %s %s\n", strings.Repeat("a", 80+i), strings.Repeat("a", 80))
					}
					return b.String(), policyWant("allow * *a*a*a*a*a*a*b *\nallow * * *a*a*a*a*a*a*b\n", b.String())
				}),
				Explain: "A recursive matcher tries every way to share 80 a's among six stars: about 20 seconds for each request, measured in C. The greedy matcher that remembers only the last star answers in well under a microsecond."},
			{Name: "a rule that is not allow or deny", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "allow * * *\npermit user:ana * *\n"}, Stdin: "user:ana read x\n", WantExit: 1, Stderr: true,
				Explain: "A typo in a policy must stop the program, not be skipped: a skipped deny is a silent hole."},
			{Name: "a rule with two patterns", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "deny user:intern storage:*\n"}, Stdin: "user:intern storage:Get x\n", WantExit: 1, Stderr: true},
			{Name: "a request with two fields", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": "allow * * *\n"}, Stdin: "user:ana read\n", WantExit: 1, Stderr: true},
			{Name: "a big policy", Args: []string{"policy.txt"}, Files: map[string]string{"policy.txt": policyBigRules()}, Gen: generated(policyBigCase),
				Explain: "2,000 rules and 5,000 requests: ten million pattern matches, which the greedy matcher handles in about a second."},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Read the rules into an array of structs (an int for allow or deny, the line number and three strings) before reading any request.",
			"sscanf(line, \"%15s %255s %255s %255s %c\", ...) returning exactly 4 is a quick test for the right number of fields (the %c catches a fifth).",
			"match(const char *p, const char *s) with the star-and-mark loop is a dozen lines; test it alone first, on *ab against aab and on the empty text.",
		},
	},
}

const totpVectors = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 59\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 1111111109\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 1111111111\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 1234567890\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 2000000000\nGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ 20000000000\n"

// totpCode is RFC 6238's code for a decoded key and a step.
func totpCode(key []byte, step uint64) uint32 {
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], step)
	m := hmac.New(sha1.New, key)
	m.Write(msg[:])
	mac := m.Sum(nil)
	o := mac[19] & 15
	return (binary.BigEndian.Uint32(mac[o:o+4]) & 0x7fffffff) % 1000000
}

// totpKey decodes Base32 by hand, as the task does: 5 bits a character,
// a byte whenever 8 are waiting, leftover bits dropped.
func totpKey(secret string) []byte {
	var out []byte
	bits, n := uint(0), 0
	for _, c := range strings.ToUpper(strings.TrimRight(secret, "=")) {
		bits = bits<<5 | uint(strings.IndexRune("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567", c))
		n += 5
		if n >= 8 {
			out = append(out, byte(bits>>(n-8)))
			n -= 8
		}
	}
	return out
}

// totpWant answers valid lines as the task describes.
func totpWant(in string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSpace(in), "\n") {
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		key := totpKey(f[0])
		t, _ := strconv.ParseUint(f[1], 10, 64)
		step := t / 30
		if len(f) == 2 {
			fmt.Fprintf(&b, "%06d\n", totpCode(key, step))
			continue
		}
		c, _ := strconv.ParseUint(f[2], 10, 32) // not Sscan: 012345 would be octal
		code := uint32(c)
		answer := "no"
		for _, d := range []int{0, -1, 1} {
			if d == -1 && step == 0 {
				continue
			}
			if totpCode(key, uint64(int64(step)+int64(d))) == code {
				answer = map[int]string{0: "ok 0", -1: "ok -1", 1: "ok +1"}[d]
				break
			}
		}
		b.WriteString(answer + "\n")
	}
	return b.String()
}

// totpSecret is n random Base32 characters.
func totpSecret(n int, seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	b := make([]byte, n)
	for i := range b {
		b[i] = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"[rng.Intn(32)]
	}
	return string(b)
}

// totpTopBitCase lists times whose truncated 4 bytes have the top bit set.
func totpTopBitCase() (string, string) {
	const secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	key := totpKey(secret)
	var b strings.Builder
	for step, found := uint64(56666666), 0; found < 5; step++ {
		var msg [8]byte
		binary.BigEndian.PutUint64(msg[:], step)
		m := hmac.New(sha1.New, key)
		m.Write(msg[:])
		mac := m.Sum(nil)
		if mac[mac[19]&15]&0x80 != 0 {
			fmt.Fprintf(&b, "%s %d\n", secret, step*30+7)
			found++
		}
	}
	return b.String(), totpWant(b.String())
}

// totpWindowCase checks codes from steps -2 to +2 around one time.
func totpWindowCase() (string, string) {
	secret := totpSecret(32, 4)
	key := totpKey(secret)
	const t = 1700000015
	step := uint64(t / 30)
	var b strings.Builder
	for _, d := range []int64{0, -1, 1, -2, 2} {
		fmt.Fprintf(&b, "%s %d %06d\n", secret, t, totpCode(key, uint64(int64(step)+d)))
	}
	fmt.Fprintf(&b, "%s %d %06d\n", secret, t, (totpCode(key, step)+1)%1000000)
	fmt.Fprintf(&b, "%s 10 %06d\n", secret, totpCode(key, 0))
	return b.String(), totpWant(b.String())
}

func totpManyCase() (string, string) {
	rng := rand.New(rand.NewSource(5))
	var b strings.Builder
	for i := 0; i < 50000; i++ {
		secret := totpSecret(16+rng.Intn(40), int64(i))
		t := rng.Int63n(4000000000)
		if rng.Intn(3) == 0 {
			fmt.Fprintf(&b, "%s %d %06d\n", secret, t, rng.Intn(1000000))
		} else {
			fmt.Fprintf(&b, "%s %d\n", secret, t)
		}
	}
	return b.String(), totpWant(b.String())
}

const (
	jwtSecret   = "my-secret"
	jwtIssuer   = "https://login.example"
	jwtAudience = "notes-app"
	jwtNow      = 1800000000
	jwtExp      = 1800000900
	jwtHeader   = `{"alg":"HS256","typ":"JWT"}`
)

func jwtArgs(now int) []string {
	return []string{jwtSecret, jwtIssuer, jwtAudience, fmt.Sprint(now)}
}

func jwtClaims(sub string, exp int, extra string) string {
	return fmt.Sprintf(`{"sub":"%s","iss":"%s","aud":"%s","iat":1799999000,"exp":%d%s}`, sub, jwtIssuer, jwtAudience, exp, extra)
}

func jwtPart(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }

func jwtSignWith(secret, header, payload string) string {
	signing := jwtPart(header) + "." + jwtPart(payload)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(signing))
	return signing + "." + base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

func jwtSign(header, payload string) string { return jwtSignWith(jwtSecret, header, payload) }

// jwtURLCharsCase finds a genuine token whose signature uses both - and _.
func jwtURLCharsCase() (string, string) {
	for i := 0; ; i++ {
		sub := fmt.Sprintf("user-%d", i)
		tok := jwtSign(jwtHeader, jwtClaims(sub, jwtExp, ""))
		sig := tok[strings.LastIndexByte(tok, '.')+1:]
		if strings.ContainsRune(sig, '-') && strings.ContainsRune(sig, '_') {
			return tok + "\n", fmt.Sprintf("valid sub=%s exp=%d\n", sub, jwtExp)
		}
	}
}

const policyStarter = "# who may do what\nallow user:* storage:Get* reports/*\ndeny  user:intern storage:* reports/secret/*\n"

const policyWildcards = "allow * logs:Read logs/*\nallow svc:*-prod deploy:* app/*/release/*\nallow * cache:* *ab\ndeny svc:test-prod deploy:Rollback *\n"

const policyWildcardRequests = "svc:x logs:Read logs/\nsvc:x logs:Read logs\nsvc:api-prod deploy:Start app/web/release/v2\nsvc:api-prod deploy:Start app/web/v2\nsvc:-prod deploy:Start app//release/\nu cache:Get aab\nu cache:Get abab\nu cache:Get aba\nsvc:test-prod deploy:Rollback app/x/release/1\n"

// policyMatch is the greedy * matcher.
func policyMatch(p, s string) bool {
	star, mark := -1, 0
	i, j := 0, 0
	for j < len(s) {
		switch {
		case i < len(p) && p[i] == '*':
			star, mark = i, j
			i++
		case i < len(p) && p[i] == s[j]:
			i++
			j++
		case star >= 0:
			i = star + 1
			mark++
			j = mark
		default:
			return false
		}
	}
	for i < len(p) && p[i] == '*' {
		i++
	}
	return i == len(p)
}

// policyWant evaluates valid requests against a valid policy.
func policyWant(policy, requests string) string {
	type rule struct {
		allow bool
		line  int
		pat   []string
	}
	var rules []rule
	for i, line := range strings.Split(policy, "\n") {
		f := strings.Fields(line)
		if len(f) == 0 || strings.HasPrefix(f[0], "#") {
			continue
		}
		rules = append(rules, rule{f[0] == "allow", i + 1, f[1:]})
	}
	var b strings.Builder
	for _, req := range strings.Split(strings.TrimSuffix(requests, "\n"), "\n") {
		f := strings.Fields(req)
		matches := func(r rule) bool {
			return policyMatch(r.pat[0], f[0]) && policyMatch(r.pat[1], f[1]) && policyMatch(r.pat[2], f[2])
		}
		answer := "deny (no rule)"
		for _, r := range rules {
			if !r.allow && matches(r) {
				answer = fmt.Sprintf("deny (line %d)", r.line)
				break
			}
		}
		if answer == "deny (no rule)" {
			for _, r := range rules {
				if r.allow && matches(r) {
					answer = fmt.Sprintf("allow (line %d)", r.line)
					break
				}
			}
		}
		b.WriteString(answer + "\n")
	}
	return b.String()
}

// policyBigRules is 2,000 generated rules; built once at startup, so it
// is kept small (about 90 KB).
func policyBigRules() string {
	rng := rand.New(rand.NewSource(71))
	services := []string{"storage", "iam", "queue", "db", "logs", "compute"}
	verbs := []string{"Get", "Put", "Delete", "List", "Create", "Update"}
	var b strings.Builder
	b.WriteString("# generated: team access rules\n")
	for i := 0; i < 2000; i++ {
		effect := "allow"
		if rng.Intn(8) == 0 {
			effect = "deny"
		}
		who := fmt.Sprintf("team:%d/*", rng.Intn(50))
		if rng.Intn(10) == 0 {
			who = "*"
		}
		action := services[rng.Intn(len(services))] + ":" + verbs[rng.Intn(len(verbs))] + "*"
		resource := fmt.Sprintf("proj-%d/*", rng.Intn(200))
		if rng.Intn(5) == 0 {
			resource = fmt.Sprintf("proj-%d/*/secret*", rng.Intn(200))
		}
		fmt.Fprintf(&b, "%s %s %s %s\n", effect, who, action, resource)
	}
	return b.String()
}

func policyBigCase() (string, string) {
	rng := rand.New(rand.NewSource(72))
	services := []string{"storage", "iam", "queue", "db", "logs", "compute"}
	verbs := []string{"Get", "Put", "Delete", "List", "Create", "Update"}
	var b strings.Builder
	for i := 0; i < 5000; i++ {
		fmt.Fprintf(&b, "team:%d/u%d %s:%s%s proj-%d/dir%d/%sfile%d\n", rng.Intn(50), rng.Intn(100),
			services[rng.Intn(len(services))], verbs[rng.Intn(len(verbs))], []string{"Object", "Item", "User", ""}[rng.Intn(4)],
			rng.Intn(200), rng.Intn(10), []string{"", "secret-"}[rng.Intn(2)], i)
	}
	return b.String(), policyWant(policyBigRules(), b.String())
}
