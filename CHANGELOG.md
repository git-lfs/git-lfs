# Git LFS Changelog

* Fix bug where `git lfs clean` will clean Git LFS pointers too #271 (@technoweenie)

* Better timeouts for the HTTP client #215 (@Mistobaan)

* Concurrent uploads through `git lfs push` #258 (@rubyist)

* Fix `git lfs smudge` behavior with zero-length file in `.git/lfs/objects` #267 (@technoweenie)

* Separate out pre-push hook behavior from `git lfs push` #263 (@technoweenie)

* Add diff/merge properties to .gitattributes #265 (@technoweenie)

* Respect `GIT_TERMINAL_PROMPT ` #257 (@technoweenie)

* Fix CLI progress bar output #185 (@technoweenie)

* Fail fast in `clean` and `smudge` commands when run without STDIN #264 (@technoweenie)

* Fix shell quoting in pre-push hook.  #235 (@mhagger)

* Fix progress bar output during file uploads.  #185 (@technoweenie)

* Change `remote.{name}.lfs_url` to `remote.{name}.lfsurl` #237 (@technoweenie)

* Swap `git config` order.  #245 (@technoweenie)

* New `git lfs pointer` command for generating and comparing pointers #246 (@technoweenie)

* Follow optional "href" property from git-lfs-authenticate SSH command #247 (@technoweenie)

* `.git/lfs/objects` spec clarifications: #212 (@rtyley), #244 (@technoweenie)

* man page updates: #228 (@mhagger)

* pointer spec clarifications: #246 (@technoweenie)

* Code comments for the untrack command: #225 (@thekafkaf)

## v0.5.0 (10 April, 2015)

* Initial public release
