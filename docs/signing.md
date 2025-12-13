[https://www.gpg4win.org](https://www.gpg4win.org)

- `gpg --full-generate-key`
- `gpg --armor --export <key>`
- `git config --global user.signingkey <key>`
- `git config --global gpg.program "C:\Program Files (x86)\GnuPG\bin\gpg.exe"`
- `git config --global commit.gpgsign true`
- `git rebase --exec "git commit --amend --no-edit -n -S" -i --root`
    - `git rebase --exec "git commit --amend --no-edit -n -S" -i <commit_hash>`