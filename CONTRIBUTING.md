# Contributing to the Systems & AI Academy

Thank you for helping! The academy is built for self-learners, so the best contributions make learning clearer, more accurate or more welcoming. You don't need to be an expert: a learner's eye catches confusing explanations that experts no longer notice.

## Ways to help (easiest first)

1. **Report a problem.** A typo, a broken link, a confusing explanation, or a screen that looks wrong. [Open an issue](https://github.com/pronexuscodex/learning_space/issues/new/choose); a screenshot helps.
2. **Suggest a free resource.** A book, course, paper or tool that helped you. It must be free, legal and on an official site (HTTPS).
3. **Improve a concept.** A better analogy, a clearer example, a fixed diagram, an extra exercise.
4. **Add a dictionary word.** Every programming word a beginner meets should be in the dictionary (`d` in the app).
5. **Fix a bug or build a feature.** Look for [`good first issue`](https://github.com/pronexuscodex/learning_space/labels/good%20first%20issue) and [`help wanted`](https://github.com/pronexuscodex/learning_space/labels/help%20wanted).
6. **Spread the word.** Star the repository, and share your progress card (`k` in the app) with #LearnInPublic.

## Where things live

Everything is in [`academy/`](academy), one Go package with no dependencies:

| To change… | Edit |
|---|---|
| a stage's concepts, exercises, glossary, resources or quiz | `curriculum_*.go` (search for the stage title) |
| how concepts connect, and stage prerequisites | `connections.go` |
| the dictionary | `vocab_*.go` |
| the Library's PDFs | `freeBooks` in `library.go` |
| a stage's classic reading | `classic.go` |
| tech-watch feeds | `techwatch.go` |

The [full guide](academy/README.md#source-layout) lists every file.

## Writing style

- Plain words and short sentences. Explain every term the first time, or link it to the dictionary.
- Keep the tone warm and never condescending: the reader is teaching themselves.
- Keep diagram and "Under the hood" lines at most 70 characters wide, because they are shown exactly as written.
- Links: HTTPS only, official sources, free to read.
- Security content is **defensive**: how to recognise, prevent and fix problems, practised only on legal training sites (OverTheWire, PortSwigger Academy, picoCTF and similar) and on your own machines.

## Making a change

You need [Go](https://go.dev/dl/) 1.22 or newer, and nothing else.

```sh
git clone https://github.com/pronexuscodex/learning_space
cd learning_space/academy
go build -o academy . && ./academy     # try it
gofmt -w . && go vet ./... && go test ./...   # what CI checks first
```

The tests check a lot for you: every stage is complete, every link is HTTPS, every concept has its connections and exercises, and nothing overflows a narrow terminal. If you changed a screen, also run `python3 tools/layout_check.py 40 80 120` (Linux or macOS).

Then open a pull request. Describe what you changed and why; for anything visible, add a screenshot. CI runs the tests on Linux, macOS and Windows. A maintainer will review it, usually within a few days.

Please keep each pull request to one topic: two small PRs are reviewed faster than one big one.

## Code of conduct

Be kind, assume good intent, and help newcomers. Harassment or insults of any kind are not tolerated. We follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/1/code_of_conduct/); report problems to the maintainer through a [private security advisory](https://github.com/pronexuscodex/learning_space/security/advisories/new) or an issue.

By contributing, you agree that your contribution is licensed under the project's [MIT license](LICENSE).
