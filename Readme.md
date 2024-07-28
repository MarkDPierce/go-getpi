# Go-Getpi

👋 Everyone. Welcome to 'Go-Getpi' a Golang rewrite of [orbital-sync](https://github.com/mattwebbio/orbital-sync), it self being a rewrite of [gravity-sync](https://github.com/vmstan/gravity-sync). I am not attempting to write the tool from the bottom-up as I believe the orbital-sync team have done a fine enough job with their implementation. As I am working on improving my Go skills as well as just getting familiar with the language I figured I'd give this a shot.

As of today (July 27th 2024) this is an MVP of what I have so far. The code builds, the code works. I still want to roll out a bit more CI/CD features leveraging Github as well as implement this as a CLI tool as well so you don't have to run it as a container.

## Todos

Write proper documentation
Setup Github Actions
    - Build Code
    - Test Code (write tests)
    - Build Container
        - Matrix array of platform archs.
    - ???
    - Profit.
Setup Github Releases
Implement Logging
Cleanup code 
Implement notifications
    - Gotify
    - email
    - webhook?
Implement a CLI
    Possible Commands/flags
        - Download Backup
        - Upload Backup
        - Reload Gravity

# Contribution

Currently, due to the maturity of this project as well as it mostly being a platform for me to learn Golang. I don't plan on accepting or implementing suggestions or code changes for the time being. 

This is not a hard no, it's more of a "Let me get bored learning, then open it up".