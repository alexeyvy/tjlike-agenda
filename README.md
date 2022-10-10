# TJ-like Agenda

<!-- TOC -->
* [Overview](#overview)
* [Installation](#installation)
* [Usage](#usage)
* [TODOs](#todos)
* [Contribution](#contribution)
* [License](#license)
<!-- TOC -->
## Overview

Pull Telegram (more platforms to be added) publications and pick up the most trending ones.

This application allows you to collect publications from pre-configured telegram (ATM) channels with no authentication required, neither a user account nor a service key. Scraping is currently done through the publicly available UI and is mainly intended to pull posts' IDs that can be used in a widget rather than downloading all the contents of the post, though it could be reworked in the future. 

Scope of application might primarily be social news aggregators such as TJournal (good night, sweet prince!) and the like which could make use of this software as a self-hosted microservice.

## Installation
- [Download](https://github.com/alexeyvy/tjlike-agenda/releases/latest) the latest archive with the binary for your OS/CPU
- Extract the archive to any folder and navigate to it
- Make sure [config.yaml](config.yaml) is placed in the working directory before running binary
- Alter `config.yaml` so that it reflects your preferred channel pool to gather publications from

## Usage
- Run the binary `./tjlike-agenda` and give it few moments to scrape publications (watch the output to track its progress)
- Upon at least 1 traversal completes, you can find the collected reposts JSON-serialized in the DB which by default is simply a file in the working directory that is called `tjlike_agenda_db.txt` and spawned/appended automatically
- However, you don't want dealing with the raw DB file which may further be replaced with another storage implementation, instead use the REST API endpoint described in [api.yml](api.yml), as follows:
```
curl localhost:35971/reposts
```
- Note that this endpoint in the current implementation isn't idempotent meaning all returned reposts evaporate from the DB as soon as the endpoint is called 
## TODOs
- Dockerize
- Better strategy on cross-channel selection
- Channel priorities
- Allow to opt for not purging read reposts
- Introduce webhooks or a queue transport to deliver reposts so periodical pulling is eliminated (?)

## Contribution
Current implementation only takes into account publication's view counter compared to previous publications' view counters. The more the counter deviates from preceding ones, the more trending it's recognized as trending within the channel which is pretty straightforward.
Whatever insights you have as to how to improve the algorithm, feel free to contribute or discuss.

Contribution to other parts are also welcome.


## License
[MIT](LICENSE)