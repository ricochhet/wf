# start the ssh-agent in the background
Get-Service -Name ssh-agent | Set-Service -StartupType Manual
Start-Service ssh-agent

ssh-add c:/Users/YOU/.ssh/id_ed25519
cat ~/.ssh/id_ed25519.pub | clip