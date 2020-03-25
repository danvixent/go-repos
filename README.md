# go-repos

A simple CLI to search a user's GitHub repositories.

Usage:  

    $ go-repos danvixent -search -must -name=go-repos -lang=go -date=2020-02-11 -desc=CLI

    OR

    $ go-repos danvixent -search -must -name go-repos -lang go -date 2020-02-11 -desc CLI
| Argument/Flag | Usage                                                                                                                |
|---------------|----------------------------------------------------------------------------------------------------------------------|
| danvixent     | Username to search GitHub for                                                                                        |
| -search       | Search will be done, if absent, all repositories of the user will be displayed                                       |
| -must         | Print only Results that match all criteria, if absent, Repositories matching at least one criteria will be displayed |
| -name         | Name Of Repository to search for                                                                                     |
| -lang         | Language Base of Repository to Search for                                                                            |
| -date         | Creation Date Of Repository to Search For                                                                            |
| -desc         | Repository Description to Search For                                                                                 |
