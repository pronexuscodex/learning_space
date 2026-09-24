#!/bin/sh
# GIT -- a whole history in a scratch folder. Run it in an empty directory.
set -e
git init -q history && cd history
git config user.name "Type In" && git config user.email "typein@example.com"
echo "first line" > notes.txt
git add notes.txt && git commit -q -m "Add notes"
echo "second line" >> notes.txt
git commit -q -am "Extend notes"
git switch -q -c idea
echo "an experiment" >> notes.txt
git commit -q -am "Try an idea"
git switch -q -
echo "commits on the main line:  $(git rev-list --count HEAD)"
echo "commits on the idea branch: $(git rev-list --count idea)"
echo "lines in notes.txt here:    $(wc -l < notes.txt | tr -d ' ')"
