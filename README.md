# website built with golang

go-website that has a backend to manage reverse proxy connections

## What is this?

This is a website built with golang. When accessing / on the given port, it will function like a normal site.
There is also an email functionality that requires a gmail app password associated with the account `noreplymasongarten`

However, it is also a [reverse proxy manager.](https://en.wikipedia.org/wiki/Reverse_proxy#:~:text=In%20computer%20networks%2C%20a%20reverse,%2C%20performance%2C%20resilience%20and%20security.)
The goal is to have an easy-to-use interface that users can delicate certain url's to act as a reverse proxy.

The proxy manager looks like this:
[Looks like this](assets/images/manager.png?raw=true "Manager")

## Install

The install is easy. Just follow below steps.

### Prerequisites:
    * golang installed
    * git installed

### Steps: 

    * `git clone https://github.com/Masong19hippows/go-website.git`

    * `cd go-website`

    * `go build`

### How to use: 

It can executed like a normal program with `.\go-website.exe` for windows and `.\go-website` for linux

There are two flags. One controls the port that the webserver uses, and the other one controls that app-password to send emails.

The port can be changed with the flag `-port <portnumber>`<br/>
The default is port 80

The gmail app-password can be changed with the flag `-password <apppassword>`<br/>
The default is nothing

# Thats All. Have Fun!!!