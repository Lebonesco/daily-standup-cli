# Daily Standup Helper

## Run Project

```
git clone https://github.com/Lebonesco/daily-standup-cli
cd daily-standup-cli
go run main.go -u Lebonesco
```

## To make this command runnable from anywhere

```bash 
go build main.go
cp main /usr/local/bin/dshelper
```

Now 

```bash
dshelper -d $HOME -u Lebonesco -a 2019-09-01
```

