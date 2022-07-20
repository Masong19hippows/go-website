# Website Built With Golang

go-website that has a backend to manage reverse proxy connections

## What is This?

This is a website built with golang. When accessing / on the given port, it will function like a normal site.
There is also an email functionality that requires a gmail app password associated with the account `noreplymasongarten`

It is also a [reverse proxy manager.](https://en.wikipedia.org/wiki/Reverse_proxy#:~:text=In%20computer%20networks%2C%20a%20reverse,%2C%20performance%2C%20resilience%20and%20security.)
The goal is to have an easy-to-use interface that users can delicate certain url's with, to act as a reverse proxy.

The proxy manager looks like this:
![Looks like this](assets/images/manager.png?raw=true "Manager")

## Install

The install is easy. Just follow below steps.

### Prerequisites
    - golang installed
    - git installed

### Steps

    - `git clone https://github.com/Masong19hippows/go-website.git`

    - `cd go-website`

    - `go build`

## How to Use: 

It can executed like a normal program with `.\go-website.exe` for windows and `./go-website` for linux

There are two flags. One controls the port that the webserver uses, and the other controls the app-password used to send emails.

The port can be changed with the flag `-port <portNumber>`<br/>
The default is port 80

The gmail app-password can be changed with the flag `-password <appPassword>`<br/>
The default is nothing

There is a default reverse proxy being servered on port 6000<br/>
You need to access this to manage proxies
You can access this by adding `/proxy` to the url to get there

When adding a proxy, if you want to add a path to the url, you must only put the url in the url spot,
<br/>and put the path in the postfix spot.

# Thats All. Have Fun!!!