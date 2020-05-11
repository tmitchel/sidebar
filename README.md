# ![Sidebar Logo](https://github.com/tmitchel/sidebar/raw/master/public/sidebar_simple_logo.png)
[![<tmitchel>](https://circleci.com/gh/tmitchel/sidebar.svg?style=shield)](https://app.circleci.com/pipelines/github/tmitchel/sidebar) [![Go Report Card](https://goreportcard.com/badge/github.com/tmitchel/sidebar)](https://goreportcard.com/report/github.com/tmitchel/sidebar) ![license](https://img.shields.io/badge/license-MIT-green) 

An open source Slack alternative with an emphasis on handling brief, tangentially related discussions that arise and discrupt the overall narrative of a conversation.

Sidebar is designed to keep communication organized in the situation where there are multiple people working on different parts of the same project. Sidebar provides a way to handle short questions/comments that arise in a discussion without allowing them to derail the entire conversation. Simply create a sidebar channel based on the message you are responding to, continue the brief converstaion there, and then mark the channel "Resolved" when the conversation is over. This way both conversations can carry on independently and all users can check the resolution of the sidebar on it is marked "Resolved".

As an example, imagine a team working remotely on separate parts of a project and you are the project lead. You may call for the current status of all team members in the group's main channel. As the responses begin pouring in, you have a question about a team members report. Rather than sending your questions in the main channel where your discussion will get mixed in with the reports from all other members, you create a sidebar as a child of the message with this member's status. Now you can ask your questions in a dedicated channel focused on this member's status. Once all of your questions are answered, you mark the sidebar "Resolved". At the end of the day, you have a main channel with the call for status followed by the reports from each team member. You also have a sidebar channel where you (and any other user) can see the more detailed discussion centered on this member's status.

### To-Do
- [ ] Allow users to deploy their own instance
- [ ] Add workspaces like Slack
- [ ] Add private channels
- [ ] File upload
- [ ] Better alerts (including mute)
- [ ] Make @ functional
- [ ] Add roles for users

## Contributing

Contributions are _greatly_ appreciated. Feel free to create and issue or submit a pull request. There are very likely a plethora of issues that need addressing.

## Installation
```
git clone git@github.com:tmitchel/sidebar.git
```

Start docker container with Postgres database
```
docker-compose -f scripts/create_db.sql up -d
```

Compile and run
```
go build -o sidebar -v cmd/chat/main.go && ./sidebar
```

## Credits

Garrett Dyson - logo - [Garrett Dyson Desgin](https://garrettdysondesign.com/)

