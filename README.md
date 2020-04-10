# go-repos

A simple CLI to search a user's GitHub repositories.

Usage:  

    $ go-repos danvixent -must -name go-repos -lang go -date 2020 -desc CLI -stars 0

        GitHub User danvixent has 1 matching Repositories:

        Respository Name    Description                                           Language  Creation Date
        -----               ------                                                ------    ------
        danvixent/go-repos  A simple CLI to search a user's GitHub repositories.  Go        March 17, 2020 12:54:11

| Argument/Flag | Usage                                                                                                                |
|---------------|----------------------------------------------------------------------------------------------------------------------|
| danvixent     | Username to search GitHub for                                                                                        |
| -must         | Print only Results that match all criteria, if absent, Repositories matching at least one criteria will be displayed |
| -name         | Name Of Repository to search for                                                                                     |
| -lang         | Language Base of Repository to Search for                                                                            |
| -date         | Creation Date Of Repository to Search For                                                                            |
| -desc         | Repository Description to Search For                                                                                 |
| -stars        | Number Of Respository Stars to search For                                     |
