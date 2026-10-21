package main

// Career content beyond coding: system design, behavioral interviews,
// resumes and the job search.

// designSteps is a framework for a 45-minute system design interview.
var designSteps = []struct{ Step, Minutes, How string }{
	{"Requirements", "5", "Functional: what users can do. Non-functional: how many users, how fast, how available, how consistent. Say what is out of scope. Ask; do not assume."},
	{"Estimates", "3", "Requests per second (a day is about 86,400 seconds, call it 10⁵), storage per day and per year, the read-to-write ratio. Round numbers are fine: you want orders of magnitude."},
	{"API", "5", "The operations with their inputs and outputs, such as POST /links {url} → {code}. This pins down what you are building."},
	{"Data model", "5", "The entities, their keys and how they are looked up, then the store that fits: SQL for relations and transactions, key-value for simple lookups at scale, object storage for files."},
	{"High-level design", "10", "Clients, load balancer, services, caches, databases, queues. Walk one request through the diagram from start to finish."},
	{"Deep dive", "10–15", "The hardest part, usually the interviewer's choice: scaling the hot path, keeping data consistent, surviving failures. Discuss options and choose one, giving reasons."},
	{"Wrap up", "3", "The bottlenecks left, the trade-offs you made, what you would monitor, and what you would build next."},
}

// designBlocks are the building blocks most designs are made from.
var designBlocks = []struct{ Name, What string }{
	{"Load balancer", "Spreads requests over many servers and stops sending to unhealthy ones."},
	{"Cache", "Keeps hot data in memory (cache-aside: read the cache, on a miss read the database and fill the cache). Hard parts: invalidation, expiry (TTL), stampedes."},
	{"CDN", "Serves static files and media from servers near the user."},
	{"SQL database", "Tables, joins, transactions and indexes. Scales reads with replicas; writes with care."},
	{"Key-value / NoSQL store", "Simple lookups by key that scale out across many machines."},
	{"Replication", "Copies of the data on several machines for safety and read capacity (leader and followers). Asynchronous copies can lag behind."},
	{"Sharding (partitioning)", "Splitting data across machines by a key, often a hash. Watch for hot keys and re-sharding."},
	{"Message queue", "Decouples producers from consumers, absorbs spikes and allows retries. Consumers must be idempotent: handling a message twice is safe."},
	{"Object storage", "Cheap, durable storage for files and blobs, addressed by key."},
	{"Rate limiter", "Caps how often a client may call (token bucket, sliding window)."},
	{"Consistent hashing", "Places keys on servers so adding or removing a server moves only a small share of the keys."},
	{"Monitoring", "Metrics, logs and traces; alerts on what users feel (SLOs), not on every blip."},
}

// latencyNumbers are rough numbers every engineer should know.
var latencyNumbers = []struct{ What, Time string }{
	{"Read from memory (RAM)", "~100 ns"},
	{"Random read from an SSD", "~100 µs"},
	{"Round trip within a data center", "~0.5 ms"},
	{"Read 1 MB sequentially from an SSD", "~1 ms"},
	{"Round trip across a continent", "~50–150 ms"},
	{"Seconds in a day", "86,400 ≈ 10⁵"},
}

// DesignPrompt is a classic system design interview question.
type DesignPrompt struct {
	Name   string
	Prompt string
	Ask    []string // clarifying questions worth asking
	Core   []string // what a strong design covers
	Deep   []string // likely deep dives
	Trap   string   // the usual mistake
	Stages []int    // where the academy teaches the pieces
}

var designPrompts = []DesignPrompt{
	{
		Name:   "URL shortener",
		Prompt: "Design a service that turns long URLs into short links and redirects visitors.",
		Ask:    []string{"Custom aliases? Expiry?", "How many new links per month, and how many redirects per link?", "Do we need click analytics?"},
		Core:   []string{"Short-code generation: a counter encoded in base 62, or a hash checked for collisions", "A key-value store from code to URL", "A redirect path served from a cache", "301 (cacheable) versus 302 redirects, which let you count every click"},
		Deep:   []string{"Generating unique codes on many servers (each server takes a block of counter values)", "Very popular links", "Abuse: rate limits and scanning for malicious URLs"},
		Trap:   "Spending the time on the database schema when reads outnumber writes 100 to 1: the cache is the design.",
		Stages: []int{7, 8},
	},
	{
		Name:   "Rate limiter",
		Prompt: "Design a rate limiter for a public API: each client may make at most N requests per minute.",
		Ask:    []string{"Limit per user, per IP address or per API key?", "One server or many?", "What should the client get when limited?"},
		Core:   []string{"An algorithm: token bucket, fixed window or sliding window, with their trade-offs", "Where it runs: in the API gateway, before the services", "Shared counters in a fast store with atomic increments", "HTTP 429 with a Retry-After header"},
		Deep:   []string{"Races between servers updating the same counter", "Bursts at the edge of a fixed window", "Failing open or closed when the counter store is down"},
		Trap:   "Keeping counters in each server's memory when there are many servers behind a load balancer.",
		Stages: []int{8, 18},
	},
	{
		Name:   "Chat app",
		Prompt: "Design a messaging app with one-to-one and group chats, online status and message history.",
		Ask:    []string{"How large can a group be?", "Delivery and read receipts?", "Message history on every device?"},
		Core:   []string{"Persistent connections (WebSockets) to connection servers", "Routing a message to the server holding the recipient's connection", "Storage partitioned by conversation, ordered by a sequence number per conversation", "Push notifications for offline users"},
		Deep:   []string{"Fan-out to large groups", "Ordering and duplicates when clients reconnect", "Presence at scale"},
		Trap:   "Polling the server every second from every client.",
		Stages: []int{8},
	},
	{
		Name:   "News feed",
		Prompt: "Design the home feed of a social network: people post, follow others and see their posts.",
		Ask:    []string{"Ranked or newest first?", "How many people does a typical user follow, and the most-followed accounts?", "Images and video?"},
		Core:   []string{"Posting service and storage", "Fan-out on write (push posts into followers' feeds) versus fan-out on read (build the feed when asked)", "A feed cache per user", "Paging with a cursor; media in object storage behind a CDN"},
		Deep:   []string{"Accounts with millions of followers: a hybrid of push and pull", "Ranking", "Deleting a post that is already in many feeds"},
		Trap:   "Choosing pure fan-out on write and forgetting celebrity accounts.",
		Stages: []int{7, 8},
	},
	{
		Name:   "Distributed key-value store",
		Prompt: "Design a key-value store that runs on many machines and survives machine failures.",
		Ask:    []string{"Consistency or availability when the network splits?", "Value sizes and the read/write mix?", "Latency targets?"},
		Core:   []string{"Partitioning with consistent hashing", "Replication: N copies, quorum reads and writes (R + W > N)", "A storage engine on each node (an LSM tree: a log, then sorted files)", "Failure detection and repairing replicas"},
		Deep:   []string{"The CAP trade-off in practice", "Conflicting writes: last-writer-wins or version vectors", "Adding a node without downtime"},
		Trap:   "Promising strong consistency and full availability at the same time.",
		Stages: []int{7, 8},
	},
	{
		Name:   "File storage and sync",
		Prompt: "Design a service like a shared drive: upload files, sync them across devices, share them.",
		Ask:    []string{"Largest file size?", "Editing on several devices at once?", "Version history?"},
		Core:   []string{"Split files into chunks, named by the hash of their content (identical chunks are stored once)", "A metadata database (files, versions, chunk lists) separate from blob storage", "Sync only the chunks that changed", "Notifying other devices of changes"},
		Deep:   []string{"Resumable uploads", "Conflicts: keep both copies and tell the user", "Sharing permissions"},
		Trap:   "Uploading the whole file again after a one-byte change.",
		Stages: []int{7, 8},
	},
	{
		Name:   "Notification system",
		Prompt: "Design a system that sends push, email and SMS notifications for many products.",
		Ask:    []string{"Which channels, and how many messages per day?", "User preferences and quiet hours?", "Must a message be sent exactly once?"},
		Core:   []string{"An API that accepts a notification request", "A queue per channel with workers calling the providers", "User preferences and templates", "Retries with backoff, and idempotency keys so a retry never sends twice"},
		Deep:   []string{"A provider outage", "Per-user rate limits so nobody is spammed", "Tracking delivery and opens"},
		Trap:   "Calling the email provider directly from the request path, so one slow provider stalls everything.",
		Stages: []int{8, 9},
	},
	{
		Name:   "Web crawler",
		Prompt: "Design a crawler that downloads a billion web pages for a search engine, and keeps them fresh.",
		Ask:    []string{"Which content types?", "How fresh must pages be?", "How polite to each website?"},
		Core:   []string{"A URL frontier: a priority queue of URLs to fetch", "Politeness: obey robots.txt and limit requests per host", "Removing duplicate URLs (normalised, with a seen-set or a Bloom filter) and duplicate content (hashes)", "Workers partitioned by host"},
		Deep:   []string{"Traps that generate endless URLs", "Choosing when to fetch a page again", "Storage for a billion pages"},
		Trap:   "Fetching from one site as fast as possible, which takes it down and gets you blocked.",
		Stages: []int{8, 12},
	},
	{
		Name:   "Metrics and logging pipeline",
		Prompt: "Design a system that collects metrics and logs from 10,000 servers, with dashboards and alerts.",
		Ask:    []string{"How long must data be kept, and at what detail?", "How quickly must an alert fire?", "How many distinct series (cardinality)?"},
		Core:   []string{"Agents on each server sending to a queue", "Stream processing to aggregate", "A time-series database, with downsampling of old data", "Alerting rules on SLOs"},
		Deep:   []string{"High-cardinality labels", "Back-pressure when the pipeline is behind", "Not losing data during an incident, when you need it most"},
		Trap:   "Alerting on every metric spike instead of on what users feel.",
		Stages: []int{19, 9},
	},
}

// BehavioralTheme is a kind of behavioral question.
type BehavioralTheme struct {
	ID, Name, Question, Tests string
}

// starGuide is how to answer a behavioral question.
var starGuide = []struct{ Part, How string }{
	{"Situation", "One or two sentences of context: where, when, what was going on."},
	{"Task", "What you were responsible for, or the goal."},
	{"Action", "Most of the answer: what you did, step by step. Say \"I\", not \"we\": they are hiring you."},
	{"Result", "What happened, with a number if you honestly have one, and what you learned or would do differently."},
}

var behavioralThemes = []BehavioralTheme{
	{"learning", "Learning fast", "Tell me about a time you had to learn something new quickly.", "How you learn: your method, not just that you did it."},
	{"hard", "A hard problem", "What is the hardest technical problem you have solved?", "Depth: can you explain it clearly and the trade-offs you weighed?"},
	{"failure", "A mistake", "Tell me about a time you made a mistake or failed.", "Honesty and ownership: what you changed afterwards."},
	{"conflict", "Disagreement", "Tell me about a time you disagreed with a teammate.", "Respect, listening and using evidence; how it was resolved."},
	{"ownership", "Ownership", "Tell me about something you improved without being asked.", "Initiative: seeing a problem and fixing it."},
	{"ambiguity", "Unclear requirements", "Tell me about a time the requirements were unclear.", "Asking questions, making reasonable assumptions, checking in."},
	{"deadline", "A tight deadline", "Tell me about delivering under a tight deadline. What did you cut?", "Prioritising and communicating trade-offs."},
	{"feedback", "Feedback", "Tell me about feedback that changed how you work.", "Openness to feedback and growth."},
	{"proud", "Proudest project", "Which project are you most proud of, and why?", "Passion and depth; expect follow-up questions on any detail."},
	{"debugging", "A hard bug", "Walk me through a difficult bug you tracked down.", "A systematic method: hypotheses, evidence, tools."},
	{"helping", "Helping others", "Tell me about a time you helped a teammate succeed.", "Collaboration and generosity."},
	{"why", "Why us", "Why do you want this role at this company?", "That you researched them; prepare one for each company."},
}

// themeName returns a theme's display name.
func themeName(id string) string {
	for _, t := range behavioralThemes {
		if t.ID == id {
			return t.Name
		}
	}
	return "General"
}

// behavioralTips are for learners who think they have no stories.
var behavioralTips = []string{
	"No job experience yet? Stories can come from your projects here, school, open source, volunteering, sport, or a job in another field. The skills are the same.",
	"Aim for about two minutes per answer. Practise aloud; writing the story first (below) makes it easy.",
	"One story can answer several questions: a hard bug is also learning fast and persistence.",
	"Have six to eight stories ready and you can answer almost any behavioral question.",
}

// resumeRules are how to write a resume engineers get interviews with.
var resumeRules = []string{
	"One page while you have under about five years of experience.",
	"Lead each bullet with a strong verb: Built, Designed, Reduced, Automated, Fixed. Never \"responsible for\".",
	"The formula: what you built + how (the technology and the key idea) + the result, with a number when you honestly have one.",
	"A Projects section with links to the code and a short demo or write-up; for career changers it matters more than education.",
	"A Skills section listing only what you can be interviewed on.",
	"Match the words of the job posting where they truly describe you: people and software screen for them.",
	"A simple single-column layout saved as a PDF; no photos, tables or graphics that screening software cannot read.",
	"Read it aloud, then have someone else read it. Typos cost interviews.",
}

// resumeExamples show weak and strong bullets.
var resumeExamples = []struct{ Weak, Strong string }{
	{"Worked on a key-value store project.",
		"Built a log-structured key-value store in C with an in-memory hash index and crash recovery; tests show no data loss across 1,000 simulated crashes."},
	{"Made a web server.",
		"Wrote an HTTP/1.1 server on epoll that serves 20,000 requests per second on a laptop, measured with wrk; documented the event-loop design."},
	{"Did security stuff for a login page.",
		"Built a login service with Argon2id password hashing, rate limiting and TOTP two-factor; added SAST checks to CI that block risky changes."},
}

// jobSearchTips are how to run a job search.
var jobSearchTips = []struct{ Title, How string }{
	{"Where the jobs are", "Company career pages, job boards, your school's career center, local meetups and conferences, open-source communities. Many roles are filled through referrals."},
	{"Referrals", "Ask people who know your work. Send a short note: the job link, your resume and two sentences on why you fit. Make it easy to say yes."},
	{"Tailor, don't spray", "A few applications adapted to each posting beat dozens of identical ones. Lead with the projects that match the role."},
	{"Keep a steady pipeline", "Apply a few times every week and track everything here. Silence is normal; it is not a judgement of you."},
	{"Follow up", "If you hear nothing after about a week of a screen, send a short, polite note."},
	{"Rejection", "Everyone gets many. Ask for feedback when you can, note what to practise, and keep the pipeline full."},
	{"Offers", "Ask for the offer in writing and for time to decide. Compare the whole package and the growth you will get. Negotiating politely, with reasons, is normal."},
}
