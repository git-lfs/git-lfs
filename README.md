GitHub Asset Management System
==============================

Large file asset management system for GitHub.

CLI
---

git-media setup
 - setup the git-media files

git-media clean
 - SHA file, see if it has changed, save to file if so, output SHA and instructions

git-media smudge
 - copy file from media center if SHA exists, (otherwise error / queue up download?)

git-media track *.jpg

- add file path or glob to .gitassets (and .gitattributes) and initial 'git add'

git-media sync

- initialize with defaults if it doesn't exist
- finds untracked large files or large files in the index
  - adds entry to .gitassets
  - adds filter to .gitattributes
  - remove from the db after (?)
- scans for changed assets and sends/checks for lock on server
- downloads assets you don't have yet, but need for your current checkout
  - re-smudge file if it was smudged without content
- uploads assets you haven't uploaded yet but have committed

git-media status

- lists all asset files in HEAD and upload/download/lock status
  - estimated file transfer size for next 'sync' command
  - only list files with interesting statuses unless --full specified

git-media view [file]

- shows version history and comments for a file

pre-commit hook that checks for large files, exits non-0 if any are too large
warn if something is committed/pushed and not uploaded
  - or post-commit to run `git-media sync` (auto-sync setting)

.gitassets
  - project id
  - large file size min limit
  - file type filter
  - list of assets and SHAs

git-media convert

- take an existing git repo and filter-branch the large files out

SERVER
------

- accept/send file contents
  - project id, path
  - full or binary diff (with base SHA)
- show list of assets in a project
- asset:
  - version history
  - visual diff if possible (cc render team)
  - locks
  - download
  - revert in project (?)
    - or create a new branch with this version
  - share (?)
    - send to campfire
    - email to someone
- keep / check for locks
