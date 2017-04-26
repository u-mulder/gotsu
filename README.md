# gotsu
Testing web apps with Go

# Contents
 - [Install](#install)
 - [Run](#run)
 - [Config format](#config-format)

# Install
Install app using `git clone` command

After install build app with `go build`

In app directory create a `config` folder.


# Run

App runs from command line:

  `> ./appname [parameters]`

Available command line parameters:

 - `-config` - mandatory, name of config, which is path in app's `/config` path
 - `-type` - optional, type of config file, currently allowed types (`json` and `sitemapxml`), `json` is default,
 - `-filename` - optional, custom name of config file without extension, by default it's `"conf"`
 - `-verbose` - optional, output more / less data when tests are run, available values: (`n`, `y`), `y` by default

Examples:

 - Test urls from `./config/mysite/conf.json` config file

  `> ./appname -config=mysite`

 - Test urls from `./config/myothersite/conf.xml` xml-config file

  `> ./appname -config=myothersite -type=sitemapxml`

 - Test urls from `./config/mysite/conf.json` config file and output only fail messages

  `> ./appname -config=mysite -verbose=n`

 - Test urls from `./config/mysite/urls.json` config file

  `> ./appname -config=mysite -filename=urls`

# Config format
App supports config files in `json` and `xml` formats.

`json` format is as follows:

    {
        "protocol": "https",
        "domain":   "www.mysite.com",
        "checkUrls": true,
        "urls": [
            {
                "url": "/url/here",
                "statusCode": 200,
                "skipUrlsCheck": false,
                "findElements": []
            },
            {
                "url": "/url/here",
                "statusCode": 200,
                "findElements": [
                    {
                        "def": ".some-class",
                        "countType": "eq",
                        "count": 3
                    }
                ]
            }
        ]
    }

Where:

- `protocol` can be either `http` or `https`
- `checkUrls` option tells to check all local links on pages
- `skipUrlsCheck` cancels `checkUrls` and all links on this concrete page will not be checked
- `findElements` will find on this concrete page elements defined by following objects
- `def` definition of collection objects to search (like jQuery selectors)
- `countType` - mode of comparing collection size against `count`, available modes are

    - `eq` - equals
    - `gt` - greater than
    - `gte` - greater than or equals
    - `lt` - less than
    - `lte` - less than or equals
    - `ne` - not equals

- `count` - expected collection size


`xml` format fully supports google sitemap xml, but has less oppotunities than json.
