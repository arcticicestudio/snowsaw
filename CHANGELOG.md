<p align="center"><img src="https://cdn.rawgit.com/arcticicestudio/snowsaw/develop/assets/snowsaw-banner.svg"/></p>

<p align="center"><img src="https://assets-cdn.github.com/favicon.ico" width=24 height=24/> <a href="https://github.com/arcticicestudio/snowsaw/releases/latest"><img src="https://img.shields.io/github/release/arcticicestudio/snowsaw.svg"/></a> <a href="https://github.com/arcticicestudio/snowsaw/releases/latest"><img src="https://img.shields.io/badge/pre--release---_-blue.svg"/></a> <img src="https://www.python.org/static/favicon.ico" width=24 height=24/> <img src="https://img.shields.io/badge/Python-3.5+-blue.svg"/></p>

---

# 0.2.0
*2017-01-28*
## Improvements
### Core Plugins
❯ The `hosts` option format of the [`link`](https://github.com/arcticicestudio/snowsaw#link) core plugin has been changed as a result to a bug also fixed in this version.
Example of the new format:
```json
[
  {
    "link": {
      "~/.gitconfig": {
        "hosts": {
          "archlinux-home": "gitconfig.home",
          "archlinux-work": "gitconfig.work"
        }
      }
    }
  }
]
```
Further information can be found in the bug fixes section below and in the associated issue #18 and PR #19.

### Documentation
❯ Added a project [debugging guide](https://github.com/arcticicestudio/snowsaw#debugging) for [JetBrains PyCharm](https://www.jetbrains.com/pycharm). (@arcticicestudio, 9694b523)
![](https://raw.githubusercontent.com/arcticicestudio/snowsaw/develop/assets/scrot-readme-debugging-run-configuration.png)

❯ Added a table of content for the [project README](https://github.com/arcticicestudio/snowsaw/blob/develop/README.md). (@arcticicestudio, 1bd1510c)

## Bug Fixes
### Core Plugins
❯ Fixed a bug where only the last duplicate link item in a snowblock configuration has been processed when using the
host-specific option `hosts` although if the host doesn't match the current hostname.
In some cases when the order of the link items has been changed also valid items for the current host have been marked
as skippable instead of linking them.

This bug was caused by an internal design conflict with the builtin Python type `dict` (dictionary) that only allows
unique keys which has been broken by defining multiple link items with the same destination path.
The new `hosts` option structure allows to define any amount of hosts with their associated target path.
(@arcticicestudio, #18 / PR #19, b921c489)

### Documentation
❯ Fixed some Markdown formatting issues in the project README. (@arcticicestudio, 7d7c0104 / e318b8d7)

# 0.1.1
*2017-01-07*
## Bug Fixes
❯ Removed the unnecessary `cd "${SNOWBLOCKSDIR}"` command in the README example [`bootstrap`](https://github.com/arcticicestudio/snowsaw#create-a-bootstrap-script) script to fix the path error `./bootstrap: line 11: .snowsaw/bin/snowsaw: No such file or directory`. (@arcticicestudio, #13, 850a72b9)

❯ Fixed a relative path mismatch error when searching for snowblock configuration files although the path must actually be absolute which caused all snowblocks to be skipped since no `snowblock.json` has been found relative to the working directory. (@arcticicestudio, #14, 4455d20f)

# 0.1.0
*2017-01-07*
## Features
❯ Implemented the [CLI][readme-cli] (@arcticicestudio, #7, 35584e0e) and public [Plugin API][readme-plugin-api] (@arcticicestudio, #6, 7bee974a).

❯ Implemented the snowsaw core logic classes
  - [`snowsaw.ConfigReader`](https://github.com/arcticicestudio/snowsaw/blob/develop/snowsaw/config.py) (@arcticicestudio, #1, bc9468df)
  - [`snowsaw.Context`](https://github.com/arcticicestudio/snowsaw/blob/develop/snowsaw/context.py) (@arcticicestudio, #2, 528d1710)
  - [`snowsaw.Dispatcher`](https://github.com/arcticicestudio/snowsaw/blob/develop/snowsaw/dispatcher.py) (@arcticicestudio, #5, 5bb0873a)

This includes the custom logger class [`snowsaw.logging.Logger`](https://github.com/arcticicestudio/snowsaw/blob/develop/snowsaw/logging/logger.py) (@arcticicestudio, #3, c56a7195) and the [`util`](https://github.com/arcticicestudio/snowsaw/tree/develop/snowsaw/util) (@arcticicestudio, #4, 695f1fd3) package which provides project util methods and classes.

❯ Implemented the [`setup.py`](https://github.com/arcticicestudio/snowsaw/blob/develop/snowsaw/setup.py) file. (@arcticicestudio, #8, 4fad0759)

❯ Implemented the core plugins
  - [Clean][readme-core-tasks-clean] (@arcticicestudio, #9, 7fa022fd)
  - [Link][readme-core-tasks-link] (@arcticicestudio, #10, 0cfd0b94)
  - [Shell][readme-core-tasks-shell] (@arcticicestudio, #11, a51b61ba)

❯ Implemented the main snowsaw executeable binary [`bin/snowsaw`](https://github.com/arcticicestudio/snowsaw/blob/develop/bin/snowsaw). (@arcticicestudio, #12, 91b9febe)

A detailed [integration guide][readme-integration-guide] and more information about the public [Plugin API][readme-plugin-api], the [design concept][readme-design-concept] and the [configuration documentation][readme-configuration-documentation] can be found in the [README][readme] and the [project wiki][wiki].

# 0.0.0
*2017-01-07*
**Project Initialization**

[readme]: https://github.com/arcticicestudio/snowsaw/blob/develop/README.md
[readme-cli]: https://github.com/arcticicestudio/snowsaw#cli
[readme-configuration-documentation]: https://github.com/arcticicestudio/snowsaw#configuration
[readme-design-concept]: https://github.com/arcticicestudio/snowsaw#design-concept
[readme-integration-guide]: https://github.com/arcticicestudio/snowsaw#integration
[readme-plugin-api]: https://github.com/arcticicestudio/snowsaw#plugin-api
[readme-core-tasks-link]: https://github.com/arcticicestudio/snowsaw#link
[readme-core-tasks-clean]: https://github.com/arcticicestudio/snowsaw#clean
[readme-core-tasks-shell]: https://github.com/arcticicestudio/snowsaw#shell
[wiki]: https://github.com/arcticicestudio/snowsaw/wiki
