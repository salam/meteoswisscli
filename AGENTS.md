# ATOMIC COMMITS

- Keep commits atomic: commit only the files you touched and list each path explicitly. For tracked files run `git commit -m "<scoped message>" -- path/to/file1 path/to/file2`. For brand-new files, use the one-liner `git restore --staged :/ && git add "path/to/file1" "path/to/file2" && git commit -m "<scoped message>" -- path/to/file1 path/to/file2`
- Never commit secrets.
- Place shell scripts, if not necessary otherwise, in the folder ./tools/
- Create a ./RELEASE_NOTES.md file for each new feature implemented. Use user-targeted language, short and brief bullet points. Create a new section ## Release 1.2 (Mon, Jan 19 09:39) for every 4 hours. E.g.

> # Version 1.0.3 (Feb 7 2025, 17:53)
> 
> * Feature 1 [author]
> * Feature 2 [author]

where author is the abbreviation of the committing person (ma = matthias/salam)
- Before attempting to delete a file to resolve a local type/lint failure, stop and ask the user. Other agents are often editing adjacent files; deleting their work to silence an error is never acceptable without explicit approval.
- NEVER edit .env or any environment variable files—only the user may change them.
- Coordinate with other agents before removing their in-progress edits—don't revert or delete work you didn't author unless everyone agrees.
- Moving/renaming and restoring files is allowed.
- Don't do git stash or other workspace changing operations. Use only non-changing git commands such as git log that do not interfere with other concurrently running AI Claude agents.
