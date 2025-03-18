### Reimplementation in GoLang

working prototype in shell github.com/michalmaruska/git-hierarchy

Can do
* walk/dump the hierarchy in topologic order
* rebase & fetch

Missing:
* renaming
* cloning
* save & restore



## Howto build & install

./build.sh

* debianization? todo




GOPRIVATE=github.com/michalmaruska

## my random notes:
* start:
indeed
 go mod
shows the commands!


go mod init {name}

go mod init git-hierarchy


* install dependencies?
go get nullprogram.com/x/optparse

*** all deps?
go mod tidy




* Emacs:
https://stackoverflow.com/questions/51642390/installing-and-using-godef


* import
https://www.geeksforgeeks.org/import-in-golang/


* why
***  "github.com/go-git/go-git/v5" // why named git not go-git
