mkdir -p ~/.ssh
ssh-keyscan -t rsa github.com >> ~/.ssh/known_hosts

eval "$(ssh-agent -s)"
ssh-keygen -t ed25519 -C "your_email@example.com"
ssh-add "c:/Users/YOU/.ssh/id_ed25519"