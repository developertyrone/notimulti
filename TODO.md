--- SPECIFY ---

Build an application that can help me to serve as a centralized notification server for all of my applications, use REST as the interface. the notification server will support different notification provider which added from time to time, so give me the flexibility to add more type in the future. Support "Telegram" and "Email" as notification provider for now. we can create multiple instances of same notification provider. configuration will be setup by a configuration file in a specific folder, the notification server will  dynamically load/fetch and reflect the existing settings or new settings.
the notification will have a simple UI with read only funcationalilty to reflect the latest settings of the server.

--- PLAN ---

The application fronend uses Tailwind CSS with Vue 3 and Vite with minimal number of libraries, But again keep things as simple as possible. Backend uses GoLang with Gin as the api framework for performance and simplicity. If necessary metadata is stored in a local SQLite database.


Use Haiku for T028-T039 (Telegram + Email providers):

These are straightforward integrations
Clear specifications in tasks.md
Standard Go patterns
Should work very well with Haiku
Switch to Sonnet for T040-T047 (API + Advanced features):

More complex orchestration
Concurrency considerations
Better handled by Sonnet's deeper reasoning

--- FEATURE2 ---
notification log 
one click button to test the provider instance
ci build for docker repository (keep secret in github repo)
k8s + docker compose deployment


--- FEATURE3 ---
release notes and tags for each release in github, like proper software
tag image with version and integrate with github action
testing and run in docker

