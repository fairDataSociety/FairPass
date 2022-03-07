# FairPass

FairPass is a Password Manager build with FairOS on top of swarm. It uses FairOS as a dependency and uses "fairpass" pod
to save "passwords" and "notes".

It is written in Go and uses [Fyne](https://developer.fyne.io/) toolkit as a platform for building ui for desktop, mobile and others.

Multiple users can use same fairpass instance. For each user a "fairpass" pod will be created. 

For now, FairPass can store Passwords and Notes, in different document tables and make them available on multiple platforms.

** FairPass needs an accessible Bee instance to work

## Warning

FairPass is work in progress, and hasn't gone through a full security audit.

Do not expect it to be bug free reliable (yet).

## How to run

### Install fyne
```
go install fyne.io/fyne/v2/cmd/fyne@latest
```

### Run
```
go run main.go
```
## Packaging

```
fyne package -os darwin -icon myapp.png
fyne package -os linux -icon myapp.png
fyne package -os windows -icon myapp.png


fyne package -os android -appID org.fairdatasociety.fairpass -icon myapp.png 
fyne package -os ios -appID org.fairdatasociety.fairpass -icon myapp.png
```

## Contribute

- Fork and clone the repository
- Make and test changes
- Open a pull request against master branch

## Credits

- [Fyne](https://github.com/fyne-io/fyne) for the UI toolkit
- [go-password](github.com/sethvargo/go-password) for password generation

