package main

// The career center: from learning to getting hired. It names the jobs
// companies hire engineers for, what their interviews test, which stages
// and portfolio projects prepare for each, and measures how ready the
// learner is: foundations, shipped projects, interview practice and a
// job-search kit. Nothing here promises a job; it measures what hiring
// checks and says what to do next.

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Career is the learner's career-center data.
type Career struct {
	Role         string            `json:"role,omitempty"`         // target role ID
	Kit          []string          `json:"kit,omitempty"`          // job-search kit items done
	Stories      []Story           `json:"stories,omitempty"`      // behavioral story bank
	Applications []Application     `json:"applications,omitempty"` // the job-search tracker
	Mocks        []MockRecord      `json:"mocks,omitempty"`        // mock interviews taken
	Bullets      map[string]string `json:"bullets,omitempty"`      // project (lab name) → resume bullet
}

// Story is one STAR story for behavioral interviews.
type Story struct {
	Theme     string    `json:"theme"` // a behavioralThemes ID
	Title     string    `json:"title"`
	Situation string    `json:"situation"`
	Task      string    `json:"task"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	At        time.Time `json:"at"`
}

// Application is one job application in the tracker.
type Application struct {
	Company string    `json:"company"`
	Role    string    `json:"role"`
	Stage   string    `json:"stage"` // one of appStages
	Next    string    `json:"next,omitempty"`
	Applied time.Time `json:"applied"`
	Updated time.Time `json:"updated"`
}

// MockRecord is one mock interview.
type MockRecord struct {
	Kind    string    `json:"kind"` // MockCoding, MockDesign or MockBehavioral
	Topic   string    `json:"topic"`
	At      time.Time `json:"at"`
	Minutes int       `json:"minutes"`
	Score   int       `json:"score"`           // rubric points met
	Of      int       `json:"of"`              // rubric points possible
	Hints   int       `json:"hints,omitempty"` // hints taken (coding problems)
}

// helped reports whether a coding problem was not yet solved alone: hints
// were taken, or the rubric was not fully met.
func (m MockRecord) helped() bool { return m.Hints > 0 || m.Score < m.Of }

// Mock interview kinds.
const (
	MockCoding     = "coding"
	MockDesign     = "design"
	MockBehavioral = "behavioral"
)

// appStages are the stages of an application, in order.
var appStages = []string{"Saved", "Applied", "Screen", "Assessment", "Interviews", "Offer", "Rejected", "Withdrawn"}

// Role is a job companies hire engineers for.
type Role struct {
	ID       string
	Name     string
	Does     string   // the work, day to day
	Hires    string   // who hires for it
	Titles   []string // job titles to search for
	Screens  []string // what the interviews test
	Loop     []string // a typical interview process
	Stages   []int    // the stages that prepare for it
	Projects []string // portfolio projects (blueprint names) employers value
	Keywords []string // skills to show on a resume, honestly
	Design   bool     // system design rounds are usual
	Advice   string   // role-specific advice
}

var roles = []Role{
	{
		ID: "backend", Name: "Backend Software Engineer",
		Does:  "Builds the services behind apps and websites: APIs, databases, queues and background jobs. Writes tests, reviews code, ships changes several times a week and keeps the services running.",
		Hires: "Almost every company that writes software: product companies, startups, banks, retailers, governments. The largest number of engineering openings.",
		Titles: []string{"Software Engineer", "Backend Engineer", "Software Developer", "Associate / Junior Software Engineer",
			"New Grad Software Engineer", "API Developer"},
		Screens: []string{
			"Coding: data structures and algorithms in one language you know deeply (Coding interview patterns [4] in the career center)",
			"HTTP and REST APIs: status codes, idempotency, authentication, pagination",
			"Databases: SQL queries, indexes, transactions, modelling data",
			"Concurrency basics: threads, races, locks",
			"Testing, Git, code review and debugging",
			"System design: basics at junior level, a full round from mid-level",
		},
		Loop: []string{
			"Recruiter call (15–30 min): your background, what you want, logistics",
			"Online assessment or phone screen: 1–2 coding problems in 45–90 min",
			"Interview loop (4–5 rounds): 2–3 coding, 1 system design (often from mid-level), 1 behavioral",
			"Team match and offer",
		},
		Stages:   []int{1, 2, 6, 7, 8, 9},
		Projects: []string{"Production-Ready To-Do API", "Epoll HTTP/1.1 Server", "Bitcask-Style KV Store", "Unix Shell", "Open-Source Contribution"},
		Keywords: []string{"REST APIs", "HTTP", "SQL", "PostgreSQL", "caching", "concurrency", "unit and integration testing", "CI/CD", "Docker", "Linux", "Git", "code review"},
		Design:   true,
		Advice:   "Pick one language (Go, Java, Python or C#) and know it deeply: its standard library, its error handling, how it does concurrency. Interviewers care more about clear, tested code than about knowing many languages.",
	},
	{
		ID: "systems", Name: "Systems / Infrastructure Engineer",
		Does:  "Works close to the machine: operating systems, storage engines, databases, compilers, networking and performance. Makes software faster, safer with memory and able to run at scale.",
		Hires: "Database and cloud companies, operating-system and browser teams, chip makers, trading firms, game-engine studios, anyone running large infrastructure.",
		Titles: []string{"Systems Software Engineer", "Infrastructure Engineer", "Storage Engineer", "Kernel Engineer",
			"Performance Engineer", "Embedded Software Engineer"},
		Screens: []string{
			"C, C++ or Rust: memory, pointers, ownership, undefined behavior",
			"Operating systems: processes, virtual memory, page faults, scheduling, system calls",
			"Concurrency: locks, atomics, races, deadlocks, memory ordering",
			"Networking: TCP, sockets, what happens when you type a URL",
			"Performance: caches, profiling with perf, reading assembly",
			"Low-level coding rounds, such as writing an allocator or an LRU cache",
		},
		Loop: []string{
			"Recruiter call",
			"Technical screen: coding plus systems questions (\"what happens on a page fault?\")",
			"Interview loop: 2 coding (often in C/C++), 1 systems deep dive, 1 design, 1 behavioral",
			"Offer",
		},
		Stages:   []int{0, 1, 4, 5, 6, 7, 8},
		Projects: []string{"Custom malloc", "User-Space Threads", "Tiny C Compiler to x86-64", "LSM-Tree Engine", "Raft Implementation"},
		Keywords: []string{"C", "C++", "Linux", "POSIX", "multithreading", "memory management", "performance profiling", "perf", "gdb", "TCP/IP", "file systems", "storage engines"},
		Design:   true,
		Advice:   "Systems teams hire people who can show they went deep: a project with a benchmark, a profiler screenshot and a write-up of what you changed and why it got faster is worth more than a list of technologies.",
	},
	{
		ID: "sre", Name: "Site Reliability / Platform Engineer",
		Does:  "Keeps services fast and available: automates deployments, builds monitoring and alerting, plans capacity, writes infrastructure as code and leads the response when something breaks.",
		Hires: "Any company running services at scale: SaaS companies, cloud providers, streaming, finance, e-commerce.",
		Titles: []string{"Site Reliability Engineer (SRE)", "Platform Engineer", "DevOps Engineer", "Cloud Engineer",
			"Infrastructure Engineer", "Production Engineer"},
		Screens: []string{
			"Linux and troubleshooting: \"a server is slow, what do you check?\" (CPU, memory, disk, network, logs)",
			"Networking: DNS, TCP, HTTP, TLS, load balancers",
			"Practical coding and scripting, often in Python or Go (parse logs, call an API, automate a task)",
			"Containers and orchestration basics, infrastructure as code",
			"Monitoring, SLOs and error budgets, alerting that wakes people only when it matters",
			"Incident response and system design with a reliability focus",
		},
		Loop: []string{
			"Recruiter call",
			"Technical screen: scripting plus Linux and networking questions",
			"Interview loop: practical coding, troubleshooting scenario, reliability-focused design, behavioral (incidents you handled)",
			"Offer",
		},
		Stages:   []int{6, 8, 9, 18, 19, 20},
		Projects: []string{"Dotfiles & Automation Kit", "Production-Ready To-Do API", "Central Logging Stack", "Hardened Cloud Baseline", "Tested Backups"},
		Keywords: []string{"Linux", "Bash", "Python", "Go", "Docker", "Kubernetes", "Terraform", "CI/CD", "Prometheus", "monitoring", "SLOs", "incident response", "on-call"},
		Design:   true,
		Advice:   "Run something real and keep it running: a small service with a deployment pipeline, dashboards and alerts, and a written post-mortem of an outage you caused and fixed on purpose. That is exactly the work, shown.",
	},
	{
		ID: "security", Name: "Security Engineer",
		Does:  "Finds and fixes weaknesses before attackers do: reviews code, models threats, secures identity and cloud setups, builds detections and responds to incidents.",
		Hires: "Tech companies, banks, healthcare, governments, consultancies and security vendors. Security teams often hire people who started in development, IT or a security operations center.",
		Titles: []string{"Security Engineer", "Application Security Engineer", "Cloud Security Engineer", "Detection Engineer",
			"SOC Analyst (a common first role)", "Security Analyst", "Penetration Tester"},
		Screens: []string{
			"Web vulnerabilities (OWASP Top 10): how each works and how to fix it in code",
			"Networking and protocols: TCP, DNS, HTTP, TLS",
			"Cryptography basics: hashing vs encryption, keys, signatures, password storage",
			"Secure code review: find the bug in a snippet",
			"Threat modelling a design; identity and access (least privilege)",
			"Incident scenarios and logs; scripting in Python",
		},
		Loop: []string{
			"Recruiter call",
			"Technical screen: security fundamentals and some scripting",
			"Interview loop: code review exercise, threat model or design, incident scenario, behavioral",
			"Offer",
		},
		Stages:   []int{8, 10, 17, 18, 19, 20},
		Projects: []string{"Secure Login Service", "Security Gates in CI", "Detection Pipeline", "Fuzz a C Parser", "CTF Journey", "Incident Response Kit"},
		Keywords: []string{"OWASP Top 10", "threat modeling", "secure code review", "SAST", "IAM", "cloud security", "SIEM", "detection engineering", "incident response", "cryptography", "Python"},
		Design:   false,
		Advice:   "Security hiring weighs skill you can show: CTF write-ups, a home lab with detections you built, a responsible bug report. Only ever test systems you own or have written permission to test.",
	},
	{
		ID: "ml", Name: "Machine Learning / AI Engineer",
		Does:  "Trains, evaluates and ships models: builds data pipelines, training and evaluation code, model serving and monitoring, and makes inference fast and affordable.",
		Hires: "AI labs, large tech companies, startups building AI products, and any company with data to learn from.",
		Titles: []string{"Machine Learning Engineer", "AI Engineer", "Applied Scientist", "Data Scientist",
			"MLOps Engineer", "ML Infrastructure Engineer"},
		Screens: []string{
			"Coding: data structures and algorithms, usually in Python",
			"ML fundamentals: overfitting, bias and variance, metrics, evaluation, cross-validation",
			"Math: linear algebra, probability, gradients",
			"Deep learning: backpropagation, attention, transformers",
			"ML system design: data → training → evaluation → serving → monitoring",
			"A deep dive into a project you built",
		},
		Loop: []string{
			"Recruiter call",
			"Technical screen: coding, sometimes ML questions",
			"Interview loop: coding, ML fundamentals, ML system design, project deep dive, behavioral",
			"Offer",
		},
		Stages:   []int{2, 13, 14, 15, 16},
		Projects: []string{"ML From Scratch", "End-to-End Prediction Project", "A/B Testing Toolkit", "GPT from Scratch", "Tiny LLM Inference Engine"},
		Keywords: []string{"Python", "NumPy", "PyTorch", "machine learning", "deep learning", "transformers", "evaluation metrics", "A/B testing", "model serving", "data pipelines", "CUDA"},
		Design:   true,
		Advice:   "Show the whole loop, not just a notebook: where the data came from, a baseline, how you evaluated, what failed, and how the model is served. One honest end-to-end project beats five tutorials.",
	},
}

// roleByID returns the role with the given ID.
func roleByID(id string) (Role, bool) {
	for _, r := range roles {
		if r.ID == id {
			return r, true
		}
	}
	return Role{}, false
}

// KitItem is one thing every job search needs.
type KitItem struct {
	ID, Name, How string
}

var jobKit = []KitItem{
	{"resume", "A one-page resume", "Every project bullet says what you built, with what, and the result (Resume builder [8] drafts them from your labs). A simple layout without tables or columns, so applicant tracking systems can read it; saved as a PDF."},
	{"github", "A GitHub profile that sells your work", "A profile README, and your best 3–6 projects pinned. Each project's README says what it is, why you built it, how to run it, the results and what you learned."},
	{"linkedin", "A professional profile", "LinkedIn or your country's equivalent, matching your resume, with the job title you want in the headline and your projects linked."},
	{"pitch", "A 60-second introduction", "\"Tell me about yourself\": who you are, what you have built, what you want next. Practised aloud until you can say it without notes."},
	{"targets", "A list of 20+ target companies", "Companies with openings for your role, a mix of large and small, with a note of why each interests you."},
	{"writeup", "One technical write-up", "A blog post or a detailed README explaining something you built: the problem, the design, the trade-offs, what you would change."},
	{"references", "Two references", "People who can speak to your work: a teacher, a mentor, an open-source maintainer, a former colleague. Ask them first."},
}

// Readiness is how ready the learner is for a role, in four parts, each
// from 0 to 1, and a weighted total.
type Readiness struct {
	Foundations float64 // concepts understood in the role's stages
	Portfolio   float64 // portfolio projects shipped, out of 3
	Practice    float64 // mock interviews and stories
	Kit         float64 // the job-search kit
	Total       float64

	Concepts, ConceptsTotal int
	Shipped                 []string // projects shipped (labs with tests passing or graduated)
	Building                []string // projects started
	Coding, Design, Stories int
}

// Readiness weights: what hiring checks most.
const (
	weightFoundations = 0.40
	weightPortfolio   = 0.25
	weightPractice    = 0.20
	weightKit         = 0.15

	targetProjects = 3
	targetCoding   = 10
	targetDesign   = 3
	targetStories  = 6
)

// readiness measures the learner against a role.
func (r *Registry) readiness(role Role) Readiness {
	var rd Readiness
	for _, id := range role.Stages {
		g, ok := guideFor(id)
		_, s := r.findStage(id)
		if !ok || s == nil {
			continue
		}
		for _, c := range g.Concepts {
			rd.ConceptsTotal++
			if s.hasStudied(c.Name) {
				rd.Concepts++
			}
		}
	}
	if rd.ConceptsTotal > 0 {
		rd.Foundations = float64(rd.Concepts) / float64(rd.ConceptsTotal)
	}

	wanted := map[string]bool{}
	for _, p := range role.Projects {
		wanted[strings.ToLower(p)] = true
	}
	points := 0.0
	for _, t := range r.Tracks {
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				done := l.CompilationStatus == CompileTested || l.Status == StatusGraduated
				switch {
				case done && wanted[strings.ToLower(l.Name)]:
					points++
					rd.Shipped = append(rd.Shipped, l.Name)
				case done:
					points += 0.5 // a finished project of your own choosing still counts
					rd.Shipped = append(rd.Shipped, l.Name)
				default:
					rd.Building = append(rd.Building, l.Name)
				}
			}
		}
	}
	rd.Portfolio = min(1, points/targetProjects)

	for _, m := range r.Career.Mocks {
		switch m.Kind {
		case MockCoding:
			rd.Coding++
		case MockDesign:
			rd.Design++
		}
	}
	rd.Stories = len(r.Career.Stories)
	parts := []float64{min(1, float64(rd.Coding)/targetCoding), min(1, float64(rd.Stories)/targetStories)}
	if role.Design {
		parts = append(parts, min(1, float64(rd.Design)/targetDesign))
	}
	for _, p := range parts {
		rd.Practice += p / float64(len(parts))
	}

	rd.Kit = float64(len(r.Career.Kit)) / float64(len(jobKit))
	rd.Total = weightFoundations*rd.Foundations + weightPortfolio*rd.Portfolio + weightPractice*rd.Practice + weightKit*rd.Kit
	return rd
}

// verdict says in words what the readiness means.
func (rd Readiness) verdict() string {
	switch {
	case rd.Total >= 0.8:
		return "Ready to apply widely. Keep practising while you interview."
	case rd.Total >= 0.6:
		return "Ready to start applying to roles that fit, while you close the gaps below."
	case rd.Total >= 0.3:
		return "Building. Keep learning and start a portfolio project now; it takes the longest."
	default:
		return "Early days. Every engineer started here: learn the foundations first."
	}
}

// careerSteps are the most useful things to do next for a role, weakest
// part first.
func (r *Registry) careerSteps(role Role, rd Readiness) []string {
	type gap struct {
		lack float64
		step string
	}
	var gaps []gap
	if rd.Foundations < 1 {
		step := "Learn the foundations: work through the stages for this role in the Study Hall [6]"
		for _, id := range role.Stages {
			g, ok := guideFor(id)
			_, s := r.findStage(id)
			if !ok || s == nil {
				continue
			}
			for _, c := range g.Concepts {
				if !s.hasStudied(c.Name) {
					step = fmt.Sprintf("Learn \"%s\" (Stage %d, Study Hall [6]): the next foundation for this role", c.Name, id)
					goto found
				}
			}
		}
	found:
		gaps = append(gaps, gap{weightFoundations * (1 - rd.Foundations), step})
	}
	if rd.Portfolio < 1 {
		step := "Ship a portfolio project: start one from Portfolio projects [3] and get its tests passing"
		if len(rd.Building) > 0 {
			step = fmt.Sprintf("Finish \"%s\": get its tests passing (Update progress [4] on the main menu), then write its README", rd.Building[0])
		}
		gaps = append(gaps, gap{weightPortfolio * (1 - rd.Portfolio), step})
	}
	if rd.Practice < 1 {
		step := "Do a timed coding mock interview [7]"
		switch {
		case rd.Stories < 3:
			step = "Write a STAR story for your story bank [6]: interviewers ask behavioral questions in every loop"
		case role.Design && rd.Design == 0:
			step = "Do a system design mock [7]: learn the framework in System design [5] first"
		}
		gaps = append(gaps, gap{weightPractice * (1 - rd.Practice), step})
	}
	if rd.Kit < 1 {
		for _, k := range jobKit {
			if !contains(r.Career.Kit, k.ID) {
				gaps = append(gaps, gap{weightKit * (1 - rd.Kit), "Job-search kit [9]: " + k.Name + ". " + k.How})
				break
			}
		}
	}
	sort.SliceStable(gaps, func(i, j int) bool { return gaps[i].lack > gaps[j].lack })
	var steps []string
	for _, g := range gaps {
		steps = append(steps, g.step)
	}
	return steps
}

// normalizeCareer repairs career data read from disk.
func (r *Registry) normalizeCareer() {
	c := &r.Career
	if _, ok := roleByID(c.Role); !ok {
		c.Role = ""
	}
	kit := c.Kit[:0]
	for _, id := range c.Kit {
		known := false
		for _, k := range jobKit {
			known = known || k.ID == id
		}
		if known && !contains(kit, id) {
			kit = append(kit, id)
		}
	}
	c.Kit = kit
	stories := c.Stories[:0]
	for _, s := range c.Stories {
		if strings.TrimSpace(s.Title) != "" {
			stories = append(stories, s)
		}
	}
	c.Stories = stories
	apps := c.Applications[:0]
	for _, a := range c.Applications {
		if strings.TrimSpace(a.Company) == "" {
			continue
		}
		if !contains(appStages, a.Stage) {
			a.Stage = "Applied"
		}
		apps = append(apps, a)
	}
	c.Applications = apps
	mocks := c.Mocks[:0]
	for _, m := range c.Mocks {
		if (m.Kind == MockCoding || m.Kind == MockDesign || m.Kind == MockBehavioral) && m.Of >= 0 {
			m.Score = min(max(0, m.Score), m.Of)
			m.Minutes = max(0, m.Minutes)
			mocks = append(mocks, m)
		}
	}
	c.Mocks = mocks
	for k, v := range c.Bullets {
		if strings.TrimSpace(v) == "" {
			delete(c.Bullets, k)
		}
	}
}

// careerMarkdown is the career export: role, readiness, projects with
// resume bullets, skills, stories and applications.
func (r *Registry) careerMarkdown(now time.Time) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# My engineering career kit\n\nExported from the Systems & AI Academy on %s.\n\n", now.Format("2 January 2006"))
	role, ok := roleByID(r.Career.Role)
	if ok {
		rd := r.readiness(role)
		fmt.Fprintf(&b, "## Target role: %s\n\nReadiness %s: foundations %s, portfolio %s, interview practice %s, job-search kit %s.\n\n%s\n\n",
			role.Name, percent(rd.Total), percent(rd.Foundations), percent(rd.Portfolio), percent(rd.Practice), percent(rd.Kit), rd.verdict())
		b.WriteString("Job titles to search for: " + strings.Join(role.Titles, ", ") + ".\n\n")
		b.WriteString("### Next steps\n\n")
		for _, s := range r.careerSteps(role, rd) {
			b.WriteString("- " + s + "\n")
		}
		b.WriteString("\n### Skills (only list what you can be interviewed on)\n\n" + strings.Join(role.Keywords, " · ") + "\n\n")
	}
	b.WriteString("## Projects\n\n")
	hasLabs := false
	for _, t := range r.Tracks {
		for _, s := range t.Stages {
			for _, l := range s.Labs {
				hasLabs = true
				fmt.Fprintf(&b, "### %s\n\n- Stage %d · %s · %s · %.1f hours\n", l.Name, s.ID, l.CompilationStatus, l.Status, l.HoursLogged)
				if bullet := r.Career.Bullets[l.Name]; bullet != "" {
					b.WriteString("- Resume bullet: " + bullet + "\n")
				}
				b.WriteString("\n")
			}
		}
	}
	if !hasLabs {
		b.WriteString("No projects yet: start one from the career center's Portfolio projects.\n\n")
	}
	b.WriteString("## Story bank (STAR)\n\n")
	if len(r.Career.Stories) == 0 {
		b.WriteString("No stories yet.\n\n")
	}
	for _, s := range r.Career.Stories {
		fmt.Fprintf(&b, "### %s (%s)\n\n- **Situation:** %s\n- **Task:** %s\n- **Action:** %s\n- **Result:** %s\n\n",
			s.Title, themeName(s.Theme), s.Situation, s.Task, s.Action, s.Result)
	}
	b.WriteString("## Applications\n\n")
	if len(r.Career.Applications) == 0 {
		b.WriteString("None yet.\n")
	} else {
		b.WriteString("| Company | Role | Stage | Applied | Next step |\n|---|---|---|---|---|\n")
		for _, a := range r.Career.Applications {
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n", mdCell(a.Company), mdCell(a.Role), a.Stage, a.Applied.Format("2006-01-02"), mdCell(a.Next))
		}
	}
	return b.String()
}

// mdCell makes text safe inside a Markdown table cell.
func mdCell(s string) string {
	return strings.NewReplacer("|", "\\|", "\n", " ").Replace(s)
}

// drillByTopic finds the coding problem a mock record is about.
func drillByTopic(topic string) (Pattern, Drill, bool) {
	for _, p := range codingPatterns {
		for _, d := range p.Drills {
			if p.Name+": "+d.Title == topic {
				return p, d, true
			}
		}
	}
	return Pattern{}, Drill{}, false
}

// InterviewRedo is a coding problem to do again, alone.
type InterviewRedo struct {
	Pattern Pattern
	Drill   Drill
	Last    MockRecord
}

// interviewRedos are the coding problems last solved with hints (or not
// fully), at least redoAfter ago: the same think-first redo as for
// curriculum exercises. A clean solve alone clears one.
func (r *Registry) interviewRedos(now time.Time) []InterviewRedo {
	latest := map[string]MockRecord{}
	var order []string
	for _, m := range r.Career.Mocks {
		if m.Kind != MockCoding {
			continue
		}
		if _, seen := latest[m.Topic]; !seen {
			order = append(order, m.Topic)
		}
		if prev, ok := latest[m.Topic]; !ok || !m.At.Before(prev.At) {
			latest[m.Topic] = m
		}
	}
	var due []InterviewRedo
	for _, topic := range order {
		m := latest[topic]
		if !m.helped() || now.Sub(m.At) < redoAfter {
			continue
		}
		if p, d, ok := drillByTopic(topic); ok {
			due = append(due, InterviewRedo{p, d, m})
		}
	}
	sort.SliceStable(due, func(i, j int) bool { return due[i].Last.At.Before(due[j].Last.At) })
	return due
}

// practisedPatterns are the coding patterns the learner has worked a
// problem of, in the order of the pattern list.
func (r *Registry) practisedPatterns() []Pattern {
	done := map[string]bool{}
	for _, m := range r.Career.Mocks {
		if m.Kind == MockCoding {
			if name, _, ok := strings.Cut(m.Topic, ": "); ok {
				done[name] = true
			}
		}
	}
	var out []Pattern
	for _, p := range codingPatterns {
		if done[p.Name] {
			out = append(out, p)
		}
	}
	return out
}

// CardPattern is the review-card kind for interview patterns.
const CardPattern = "interview pattern"

// patternCards are the review cards for a practised pattern: spotting it
// from the signals in a problem, and how it works. Recognising the
// pattern fast is the skill a coding interview tests most.
func patternCards(p Pattern) []Card {
	return []Card{
		{ID: "pattern :: " + p.Name + " :: spot", StageID: NoStage, Concept: p.Name, Kind: CardPattern,
			Q: "Coding interview: which pattern do these signals point to, and why? " + p.Signal,
			A: p.Name + ". " + p.Idea},
		{ID: "pattern :: " + p.Name + " :: how", StageID: NoStage, Concept: p.Name, Kind: CardPattern,
			Q: "How do you apply the “" + p.Name + "” pattern, what does it usually cost, and what is the classic mistake?",
			A: strings.Join(p.Steps, " → ") + "\n\nCost: " + p.Cost + "\n\nWatch out: " + p.Pitfall},
	}
}

// careerCards are the review cards from the career center.
func (r *Registry) careerCards() []Card {
	var cards []Card
	for _, p := range r.practisedPatterns() {
		cards = append(cards, patternCards(p)...)
	}
	return cards
}

// interviewSuggestion is the career step for What's next: a due redo of
// an interview problem, or a weekly problem once there is a target role
// and some foundations.
func (r *Registry) interviewSuggestion(now time.Time) (Suggestion, bool) {
	if redos := r.interviewRedos(now); len(redos) > 0 {
		return Suggestion{Kind: SuggestInterview,
			Title: "Redo an interview problem alone: " + redos[0].Drill.Title,
			Why:   fmt.Sprintf("On %s you solved it with help. Solving it again from a blank page is what makes the pattern yours.", redos[0].Last.At.Format("2 January"))}, true
	}
	role, ok := roleByID(r.Career.Role)
	if !ok {
		return Suggestion{}, false
	}
	if rd := r.readiness(role); rd.Foundations < 0.2 {
		return Suggestion{}, false
	}
	for _, m := range r.Career.Mocks {
		if m.Kind == MockCoding && now.Sub(m.At) < 7*24*time.Hour {
			return Suggestion{}, false
		}
	}
	return Suggestion{Kind: SuggestInterview, Title: "Practise an interview problem (" + role.Name + ")",
		Why: "One problem a week, worked aloud against the clock, keeps you ready to interview."}, true
}

// Career achievements.
func init() {
	readyFor := func(r *Registry, at float64) bool {
		role, ok := roleByID(r.Career.Role)
		return ok && r.readiness(role).Total >= at
	}
	achievements = append(achievements,
		Achievement{"target-role", "Eyes on the Prize", "Chose the job you are aiming for.",
			func(r *Registry, cs campusStats) bool { return r.Career.Role != "" }},
		Achievement{"first-mock", "Under the Clock", "Took your first mock interview.",
			func(r *Registry, cs campusStats) bool { return len(r.Career.Mocks) > 0 }},
		Achievement{"pattern-spotter", "Pattern Spotter", "Practised a problem from 8 different interview patterns.",
			func(r *Registry, cs campusStats) bool { return len(r.practisedPatterns()) >= 8 }},
		Achievement{"storyteller", "Storyteller", "Wrote six STAR stories for behavioral interviews.",
			func(r *Registry, cs campusStats) bool { return len(r.Career.Stories) >= targetStories }},
		Achievement{"shipped", "Shipped It", "Finished a portfolio project for your target role, tests passing.",
			func(r *Registry, cs campusStats) bool {
				role, ok := roleByID(r.Career.Role)
				if !ok {
					return false
				}
				for _, p := range role.Projects {
					for _, t := range r.Tracks {
						for _, s := range t.Stages {
							for _, l := range s.Labs {
								if strings.EqualFold(l.Name, p) && (l.CompilationStatus == CompileTested || l.Status == StatusGraduated) {
									return true
								}
							}
						}
					}
				}
				return false
			}},
		Achievement{"ready-to-apply", "Ready to Apply", "Reached 60% readiness for your target role.",
			func(r *Registry, cs campusStats) bool { return readyFor(r, 0.6) }},
		Achievement{"in-the-arena", "In the Arena", "Tracked your first job application.",
			func(r *Registry, cs campusStats) bool { return len(r.Career.Applications) > 0 }},
		Achievement{"offer", "Offer in Hand", "Reached the offer stage of an application.",
			func(r *Registry, cs campusStats) bool {
				for _, ap := range r.Career.Applications {
					if ap.Stage == "Offer" {
						return true
					}
				}
				return false
			}},
	)
}
