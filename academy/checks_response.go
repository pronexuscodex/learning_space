package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Checked exercises for Stage 20 (incident response and privacy):
// verifying a backup, building a timeline, and measuring k-anonymity.
var responseChecks = []Check{
	{
		ID: "backup_verify", Stage: 20, Concept: "Backups & Recovery Testing",
		Title: "A backup verifier",
		Task:  "./backup_verify MANIFEST DIR checks a restored folder against a manifest written by sha256sum: each line is 64 lowercase hex digits, two spaces, and a path relative to DIR. A path that is empty, starts with /, or has an empty, . or .. part makes the manifest invalid: print a message on standard error and exit with status 1 before reading any file. Otherwise, for each manifest line in order, print missing PATH if DIR has no regular file there, or changed PATH if its SHA-256 differs. Then print extra PATH for every regular file under DIR (in subfolders too) that the manifest does not list, in byte order of the path. Finally print ok N changed C missing M extra E, the counts of each, and exit with status 0: the last line is the verdict. A missing argument gets exit status 2.",
		File:  "backup_verify.c",
		Starter: `/* Checked exercise: backup_verify
 *   ./backup_verify manifest.txt restored
 *   changed docs/plan.txt
 *   missing photos/cat.jpg
 *   extra notes/new.txt
 *   ok 3 changed 1 missing 1 extra 1
 * The SHA-256 is the one from your Stage 17 sha256.c (read files in
 * pieces). Walk DIR with opendir and readdir; lstat tells you whether
 * an entry is a regular file (S_ISREG) or a folder (S_ISDIR), and skips
 * symbolic links. Collect the paths you find, sort them, and compare
 * with the manifest's.
 * Check every manifest path before opening anything: a manifest from a
 * damaged or tampered backup must not make you read ../../etc/shadow.
 */
#define _POSIX_C_SOURCE 200809L
#include <dirent.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"manifest.txt", "restored"}, Files: backupStarterFiles, Want: "changed docs/plan.txt\nmissing photos/cat.jpg\nextra notes/new.txt\nok 3 changed 1 missing 1 extra 1\n"},
			{Name: "a perfect restore", Args: []string{"manifest.txt", "r"}, Files: backupPerfectFiles, Want: "ok 6 changed 0 missing 0 extra 0\n"},
			{Name: "a file emptied", Args: []string{"manifest.txt", "r"}, Files: backupEmptiedFiles, Want: "changed db/orders.csv\nok 4 changed 1 missing 0 extra 0\n",
				Explain: "The file exists, with the right name, but is empty: only the hash notices. A check of names or sizes alone would have passed a backup job that wrote nothing."},
			{Name: "binary files", Args: []string{"manifest.txt", "r"}, Files: backupBinaryFiles, Want: "ok 2 changed 0 missing 0 extra 0\n",
				Explain: "Images and archives contain zero bytes: read with fread, not with string functions."},
			{Name: "extra files in order", Args: []string{"manifest.txt", "r"}, Files: backupExtraFiles, Want: "extra a/b/z.txt\nextra a/c.txt\nextra b.txt\nok 1 changed 0 missing 0 extra 3\n",
				Explain: "Files the manifest does not know about, in subfolders too, sorted by path byte by byte: a/b/z.txt comes before a/c.txt because b comes before c."},
			{Name: "a path that leaves the folder", Args: []string{"manifest.txt", "r"}, Files: map[string]string{"manifest.txt": sha256Hex("x") + "  ../secret.txt\n", "secret.txt": "x", "r/keep.txt": "k"}, WantExit: 1, Stderr: true,
				Explain: "A manifest path with .. would make the checker read outside the restored folder; refuse the whole manifest before opening anything."},
			{Name: "an absolute path", Args: []string{"manifest.txt", "r"}, Files: map[string]string{"manifest.txt": sha256Hex("x") + "  /etc/hostname\n", "r/keep.txt": "k"}, WantExit: 1, Stderr: true},
			{Name: "two hundred files", Args: []string{"manifest.txt", "r"}, Files: backupManyFiles(), Want: backupManyWant},
			{Name: "no arguments", WantExit: 2, Stderr: true},
		},
		Hints: []string{
			"Read the manifest into an array of (hash, path) first, checking each path; then hash each file with snprintf(full, sizeof full, \"%s/%s\", dir, path).",
			"A recursive walk(dir, prefix) that appends prefix/name for files and recurses into folders (skipping . and ..) collects every path.",
			"Sort the found paths with qsort and strcmp, and for each, a binary search or a linear scan of the manifest tells whether it is extra.",
		},
	},
	{
		ID: "timeline", Stage: 20, Concept: "Building an Incident Timeline",
		Title: "An incident timeline in UTC",
		Task:  "Read lines from standard input. clock HOST SECONDS says HOST's clock runs SECONDS fast (a whole number, perhaps negative); every other line is an event, HOST TIME MESSAGE, where TIME is YYYY-MM-DDTHH:MM:SS, optionally a . and 1 to 9 digits of fraction, then Z or an offset +HH:MM or -HH:MM (up to 14:00), and MESSAGE is the rest of the line. Convert each time to UTC, subtract the host's clock error (clock lines apply to all of that host's events, wherever they appear), and print every event as UTCTIME HOST MESSAGE, with UTCTIME as YYYY-MM-DDTHH:MM:SS.ffffffZ (the fraction cut to microseconds), sorted by that time; events at the same moment keep their input order. Years are 1970 to 9999. A date that does not exist (such as February 30, or February 29 in a year that is not a leap year), any other malformed line, or a second clock line for the same host gets a message on standard error and exit status 1, with nothing printed. There can be 100,000 events.",
		File:  "timeline.c",
		Starter: `/* Checked exercise: timeline
 *   input:   clock web 5
 *            fw 2026-10-24T09:41:07+02:00 allowed 203.0.113.9 -> web:443
 *            web 2026-10-24T07:41:15Z POST /login user=ana
 *   output:  2026-10-24T07:41:07.000000Z fw allowed 203.0.113.9 -> web:443
 *            2026-10-24T07:41:10.000000Z web POST /login user=ana
 * 09:41:07 at +02:00 is 07:41:07 UTC; web's clock is 5 s fast.
 * Turn each time into seconds since 1970 (UTC) plus microseconds:
 *   days_from_civil(y, m, d)  (do not use mktime: it applies the
 *   machine's own time zone), then subtract the offset and the clock error.
 * A fraction of .5 is 500000 microseconds, not 5.
 * Leap years: divisible by 4, except centuries, except every 400 years.
 * Sort with qsort on (time, input position), so that ties keep their order.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Stdin: timelineStarter, Want: timelineWant(timelineStarter)},
			{Name: "offsets both ways", Stdin: "a 2026-03-01T00:30:00+05:30 east\nb 2026-02-28T20:00:00-04:00 west\nc 2026-02-28T23:59:59Z utc\n", Want: timelineWant("a 2026-03-01T00:30:00+05:30 east\nb 2026-02-28T20:00:00-04:00 west\nc 2026-02-28T23:59:59Z utc\n"),
				Explain: "+05:30 is ahead of UTC, so subtract it; -04:00 is behind, so add it. All three land within a second of midnight, and the dates change in both directions."},
			{Name: "leap years", Stdin: "x 2024-02-29T12:00:00Z leap\nx 2000-02-29T12:00:00Z leap, a century divisible by 400\nx 2024-03-01T00:00:00+01:00 just before March in UTC\n", Want: timelineWant("x 2024-02-29T12:00:00Z leap\nx 2000-02-29T12:00:00Z leap, a century divisible by 400\nx 2024-03-01T00:00:00+01:00 just before March in UTC\n")},
			{Name: "not a leap year", Stdin: "x 2100-02-29T12:00:00Z no such day\n", WantExit: 1, Stderr: true,
				Explain: "2100 is divisible by 4 but is a century not divisible by 400: it has no February 29."},
			{Name: "fractions", Stdin: "h 2026-10-24T07:00:00.5Z half\nh 2026-10-24T07:00:00.000005Z five micro\nh 2026-10-24T07:00:00.1234567Z cut to micro\nh 2026-10-24T07:00:00Z whole\n", Want: timelineWant("h 2026-10-24T07:00:00.5Z half\nh 2026-10-24T07:00:00.000005Z five micro\nh 2026-10-24T07:00:00.1234567Z cut to micro\nh 2026-10-24T07:00:00Z whole\n"),
				Explain: ".5 is half a second: 500,000 microseconds. Reading the digits as a plain number makes it 5 microseconds and sorts it before .000005's neighbour."},
			{Name: "clock errors", Stdin: "clock slow -30\nfast 2026-10-24T08:00:40Z first, by the corrected clock\nslow 2026-10-24T07:59:50Z second, by the corrected clock\nclock fast 60\n", Want: timelineWant("clock slow -30\nfast 2026-10-24T08:00:40Z first, by the corrected clock\nslow 2026-10-24T07:59:50Z second, by the corrected clock\nclock fast 60\n"),
				Explain: "fast runs 60 s ahead, so its 08:00:40 is really 07:59:40; slow runs 30 s behind, so its 07:59:50 is really 08:00:20. Read as written they are in the opposite order. The clock line for fast comes last, and still applies."},
			{Name: "ties keep their order", Gen: generated(func() (string, string) {
				var b strings.Builder
				for i := 0; i < 50; i++ {
					fmt.Fprintf(&b, "h%d 2026-10-24T07:00:00Z event %d\n", i%3, i)
				}
				return b.String(), timelineWant(b.String())
			}),
				Explain: "Fifty events in the same second: qsort is not stable, so compare the input positions when the times are equal."},
			{Name: "a malformed time", Stdin: "x 2026-10-24 07:00:00 no T and no zone\n", WantExit: 1, Stderr: true},
			{Name: "February 30", Stdin: "x 2026-02-30T00:00:00Z\n", WantExit: 1, Stderr: true},
			{Name: "a hundred thousand events", Gen: generated(timelineManyCase)},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"days_from_civil: y -= m <= 2; era = y / 400; yoe = y - era * 400; doy = (153 * (m + (m > 2 ? -3 : 9)) + 2) / 5 + d - 1; doe = yoe * 365 + yoe / 4 - yoe / 100 + doy; return era * 146097 + doe - 719468.",
			"Keep times as long long microseconds since 1970; to print, split back into days and seconds and use civil_from_days, the same algorithm in reverse.",
			"Read every line into an array first (clock lines can come after events), then apply the clock errors, then sort, then print.",
		},
	},
	{
		ID: "k_anon", Stage: 20, Concept: "Anonymisation, Pseudonymisation & Differential Privacy",
		Title: "Measuring k-anonymity",
		Task:  "./k_anon K SPEC... reads comma-separated rows from standard input: a header line of column names, then rows with the same number of fields (no quoting). Each SPEC names a quasi-identifier column, optionally generalised: COLUMN keeps the value, COLUMN:N keeps its first N characters (or all of it, if shorter), COLUMN/W turns a whole number of 0 or more into its band of width W, written LO-HI (37 with /10 is 30-39). Each row's key is its generalised values joined by commas, in SPEC order. Print rows R, groups G (distinct keys), smallest S (the size of the smallest group), at risk A (rows in groups of fewer than K rows), then one line SIZE KEY for each group smaller than K, ordered by size and then by key in byte order. An unknown column, a malformed SPEC, a row with the wrong number of fields, or a value that /W cannot read gets a message on standard error and exit status 1, with nothing printed. There can be 100,000 rows.",
		File:  "k_anon.c",
		Starter: `/* Checked exercise: k_anon
 *   ./k_anon 2 postcode:3 age/10 sex   with
 *   name,postcode,age,sex
 *   Ann,SW1A2,34,F
 *   Bea,SW1A9,37,F
 *   Cal,N1C4,52,M
 *   output:  rows 3
 *            groups 2
 *            smallest 1
 *            at risk 1
 *            1 N1C,50-59,M
 * Ann and Bea share SW1,30-39,F: a group of 2. Cal is alone, so anyone
 * who knows his postcode area, age band and sex can find his row.
 * Build each row's key as a string, collect all keys, sort them with
 * qsort and strcmp, and count runs of equal keys. Keep the commas
 * between fields in the key: without them, "ab"+"c" and "a"+"bc" would
 * be the same group.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
`,
		Cases: []CheckCase{
			{Name: "the starter's example", Args: []string{"2", "postcode:3", "age/10", "sex"}, Stdin: kanonStarter, Want: kanonWant(kanonStarter, 2, []string{"postcode:3", "age/10", "sex"})},
			{Name: "no generalisation", Args: []string{"2", "postcode", "birth", "sex"}, Gen: generated(func() (string, string) {
				in := kanonPeople(400, 101)
				return in, kanonWant(in, 2, []string{"postcode", "birth", "sex"})
			}),
				Explain: "With the full postcode, birth date and sex, almost every one of 400 people is in a group of one: unique, and so identifiable."},
			{Name: "generalised", Args: []string{"5", "postcode:2", "birth:4", "sex"}, Gen: generated(func() (string, string) {
				in := kanonPeople(400, 101)
				return in, kanonWant(in, 5, []string{"postcode:2", "birth:4", "sex"})
			}),
				Explain: "Postcode area and birth year instead: far fewer, bigger groups. The groups still under 5 are listed so they can be generalised further or suppressed."},
			{Name: "bands", Args: []string{"3", "age/10", "income/25000"}, Stdin: "id,age,income\n1,0,0\n2,9,24999\n3,10,25000\n4,19,49999\n5,100,1000000\n", Want: kanonWant("id,age,income\n1,0,0\n2,9,24999\n3,10,25000\n4,19,49999\n5,100,1000000\n", 3, []string{"age/10", "income/25000"}),
				Explain: "Band edges: 9 is in 0-9 and 10 in 10-19; the band's top is LO + W - 1."},
			{Name: "keys that run together", Args: []string{"2", "a", "b"}, Stdin: "a,b\nab,c\na,bc\nx,y\nx,y\n", Want: kanonWant("a,b\nab,c\na,bc\nx,y\nx,y\n", 2, []string{"a", "b"}),
				Explain: "ab and c, and a and bc, are different people: joined without a comma, both keys are abc and they would look like a group of two."},
			{Name: "an unknown column", Args: []string{"2", "zip"}, Stdin: "postcode,sex\nN1,F\n", WantExit: 1, Stderr: true},
			{Name: "a row with too few fields", Args: []string{"2", "sex"}, Stdin: "postcode,sex\nN1,F\nN2\n", WantExit: 1, Stderr: true,
				Explain: "A short row means the columns are shifted: grouping it would mix up fields, so refuse the file."},
			{Name: "a value that is not a number", Args: []string{"2", "age/10"}, Stdin: "age\n34\nforty\n", WantExit: 1, Stderr: true},
			{Name: "a hundred thousand rows", Args: []string{"10", "postcode:3", "birth:4", "sex"}, Gen: generated(func() (string, string) {
				in := kanonPeople(100000, 102)
				return in, kanonWant(in, 10, []string{"postcode:3", "birth:4", "sex"})
			})},
		},
		Timeout: 10 * time.Second,
		Hints: []string{
			"Split each line at commas into an array of field pointers (strtok skips empty fields; walk with strchr instead, so a,,b has three fields).",
			"Parse each SPEC once into (column index, kind, N or W) before reading the rows.",
			"After sorting the keys, one pass counts groups; store (size, key) for the small ones and sort those with a second comparison function.",
		},
	},
}

var backupStarterFiles = map[string]string{
	"manifest.txt":             sha256Hex("plan v1\n") + "  docs/plan.txt\n" + sha256Hex("budget\n") + "  docs/budget.txt\n" + sha256Hex("meow") + "  photos/cat.jpg\n" + sha256Hex("dog") + "  photos/dog.jpg\n" + sha256Hex("readme\n") + "  readme.txt\n",
	"restored/docs/plan.txt":   "plan v2\n",
	"restored/docs/budget.txt": "budget\n",
	"restored/photos/dog.jpg":  "dog",
	"restored/readme.txt":      "readme\n",
	"restored/notes/new.txt":   "new\n",
}

var backupPerfectFiles = backupTree(map[string]string{"a.txt": "alpha\n", "b/c.txt": "charlie\n", "b/d/e.txt": "echo\n", "f.txt": "", "g/h.bin": "\x00\x01\x02", "i.txt": strings.Repeat("ii", 5000)})

var backupEmptiedFiles = func() map[string]string {
	m := backupTree(map[string]string{"db/orders.csv": "id,total\n1,9.99\n", "db/users.csv": "id\n1\n", "conf/app.toml": "x=1\n", "logo.png": "\x89PNG", "README": "r\n"})
	m["r/db/orders.csv"] = ""
	return m
}()

var backupBinaryFiles = backupTree(map[string]string{"photo.jpg": "\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01" + strings.Repeat("\x00", 300) + "end", "archive.zip": "PK\x03\x04\x00\x00\x00\x00rest"})

var backupExtraFiles = func() map[string]string {
	m := backupTree(map[string]string{"a/keep.txt": "k\n"})
	m["r/b.txt"] = "b"
	m["r/a/c.txt"] = "c"
	m["r/a/b/z.txt"] = "z"
	return m
}()

// backupTree is a manifest of files and the folder r holding exactly them.
func backupTree(files map[string]string) map[string]string {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	m := map[string]string{}
	var b strings.Builder
	for _, p := range paths {
		fmt.Fprintf(&b, "%s  %s\n", sha256Hex(files[p]), p)
		m["r/"+p] = files[p]
	}
	m["manifest.txt"] = b.String()
	return m
}

// backupManyWant is what backupManyFiles makes the verifier report.
const backupManyWant = "missing d1/f050.dat\nchanged d3/f010.dat\nextra d1/f050.dat.bak\nextra zz/new.dat\nok 198 changed 1 missing 1 extra 2\n"

// backupManyFiles is 200 small files (about 250 KB, built at startup),
// with a few changed, missing and extra.
func backupManyFiles() map[string]string {
	rng := rand.New(rand.NewSource(111))
	files := map[string]string{}
	for i := 0; i < 200; i++ {
		data := make([]byte, 200+rng.Intn(2200))
		rng.Read(data)
		files[fmt.Sprintf("d%d/f%03d.dat", i%7, i)] = string(data)
	}
	m := backupTree(files)
	m["r/d3/f010.dat"] += "x"
	delete(m, "r/d1/f050.dat")
	m["r/d1/f050.dat.bak"] = "old"
	m["r/zz/new.dat"] = "new"
	return m
}

const timelineStarter = "clock web 5\nfw 2026-10-24T09:41:07+02:00 allowed 203.0.113.9 -> web:443\nweb 2026-10-24T07:41:15Z POST /login user=ana\n"

// timelineWant builds the timeline of a valid input.
func timelineWant(in string) string {
	type event struct {
		t        time.Time
		host     string
		msg      string
		position int
	}
	skew := map[string]int{}
	var evs []event
	for i, line := range strings.Split(strings.TrimSuffix(in, "\n"), "\n") {
		f := strings.SplitN(line, " ", 3)
		if f[0] == "clock" {
			s, _ := strconv.Atoi(strings.Fields(line)[2])
			skew[f[1]] = s
			continue
		}
		t, err := time.Parse(time.RFC3339Nano, f[1])
		if err != nil {
			panic(err)
		}
		msg := ""
		if len(f) == 3 {
			msg = f[2]
		}
		evs = append(evs, event{t, f[0], msg, i})
	}
	for i := range evs {
		evs[i].t = evs[i].t.Add(-time.Duration(skew[evs[i].host]) * time.Second)
	}
	sort.SliceStable(evs, func(a, b int) bool { return evs[a].t.Before(evs[b].t) })
	var b strings.Builder
	for _, e := range evs {
		fmt.Fprintf(&b, "%s %s %s\n", e.t.UTC().Format("2006-01-02T15:04:05.000000Z"), e.host, e.msg)
	}
	return b.String()
}

func timelineManyCase() (string, string) {
	rng := rand.New(rand.NewSource(121))
	hosts := []string{"fw", "web1", "web2", "db", "idp", "laptop"}
	zones := []string{"Z", "+02:00", "-05:00", "+05:30", "Z", "Z"}
	var b strings.Builder
	b.WriteString("clock web2 -12\nclock laptop 95\n")
	base := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 100000; i++ {
		h := rng.Intn(len(hosts))
		t := base.Add(time.Duration(rng.Int63n(int64(48 * time.Hour))))
		if rng.Intn(3) > 0 {
			t = t.Truncate(time.Second)
		} else {
			t = t.Truncate(time.Millisecond)
		}
		zone := zones[h]
		var stamp string
		if zone == "Z" {
			stamp = t.Format("2006-01-02T15:04:05.999999999Z")
		} else {
			loc := time.FixedZone("", map[string]int{"+02:00": 7200, "-05:00": -18000, "+05:30": 19800}[zone])
			stamp = t.In(loc).Format("2006-01-02T15:04:05.999999999-07:00")
		}
		fmt.Fprintf(&b, "%s %s event %d\n", hosts[h], stamp, i)
	}
	return b.String(), timelineWant(b.String())
}

const kanonStarter = "name,postcode,age,sex\nAnn,SW1A2,34,F\nBea,SW1A9,37,F\nCal,N1C4,52,M\n"

// kanonWant measures k-anonymity of a valid input.
func kanonWant(in string, k int, specs []string) string {
	lines := strings.Split(strings.TrimSuffix(in, "\n"), "\n")
	header := strings.Split(lines[0], ",")
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	type spec struct{ c, n, w int }
	var ss []spec
	for _, s := range specs {
		switch {
		case strings.Contains(s, ":"):
			p := strings.SplitN(s, ":", 2)
			n, _ := strconv.Atoi(p[1])
			ss = append(ss, spec{col[p[0]], n, 0})
		case strings.Contains(s, "/"):
			p := strings.SplitN(s, "/", 2)
			w, _ := strconv.Atoi(p[1])
			ss = append(ss, spec{col[p[0]], 0, w})
		default:
			ss = append(ss, spec{col[s], 0, 0})
		}
	}
	count := map[string]int{}
	for _, line := range lines[1:] {
		f := strings.Split(line, ",")
		var parts []string
		for _, s := range ss {
			v := f[s.c]
			switch {
			case s.n > 0 && len(v) > s.n:
				v = v[:s.n]
			case s.w > 0:
				x, _ := strconv.Atoi(v)
				lo := x / s.w * s.w
				v = fmt.Sprintf("%d-%d", lo, lo+s.w-1)
			}
			parts = append(parts, v)
		}
		count[strings.Join(parts, ",")]++
	}
	type group struct {
		size int
		key  string
	}
	var small []group
	smallest, atRisk := -1, 0
	for key, n := range count {
		if smallest < 0 || n < smallest {
			smallest = n
		}
		if n < k {
			small = append(small, group{n, key})
			atRisk += n
		}
	}
	sort.Slice(small, func(a, b int) bool {
		if small[a].size != small[b].size {
			return small[a].size < small[b].size
		}
		return small[a].key < small[b].key
	})
	var b strings.Builder
	fmt.Fprintf(&b, "rows %d\ngroups %d\nsmallest %d\nat risk %d\n", len(lines)-1, len(count), smallest, atRisk)
	for _, g := range small {
		fmt.Fprintf(&b, "%d %s\n", g.size, g.key)
	}
	return b.String()
}

// kanonPeople is n simulated people with a postcode, birth date and sex.
func kanonPeople(n int, seed int64) string {
	rng := rand.New(rand.NewSource(seed))
	var b strings.Builder
	b.WriteString("id,postcode,birth,sex,visits\n")
	for i := 0; i < n; i++ {
		area := []string{"N1", "N7", "E2", "SW", "SE", "W4", "NW", "EC"}[rng.Intn(8)]
		day := time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, rng.Intn(60*365))
		fmt.Fprintf(&b, "%d,%s%d%c,%s,%s,%d\n", i, area, rng.Intn(10), 'A'+rune(rng.Intn(5)), day.Format("2006-01-02"), []string{"F", "M"}[rng.Intn(2)], rng.Intn(50))
	}
	return b.String()
}
