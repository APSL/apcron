# apcron

Cron for applications, with modern logging and monitoring features. Written in GO.

[![Go Report Card](https://goreportcard.com/badge/github.com/apsl/go-cron)](https://goreportcard.com/report/github.com/apsl/go-cron)

## Overview

Simple app-oriented cron daemon with these features:

* Single and little binary executable with no dependences that can be copied into your container, or managed in your server with systemd.
* Statistical data about executions. Memory, execution time, number of executions.
* Real-time logs. Log output as your app writes stdout/stderr.
* Colorized log oputput to facilitate distinction of errors and info when console is detected.
* Sentry enabled, for sending  app startup errors, stderr output, or even log errors.
* Developer-oriented crontab specified in yaml format.
* Ability to launch through shell or direct process. Configurable for each command. 
* Low memory consumption.

It uses [robfig/cron](https://github.com/robfig/cron) for golang process
scheduling. The [manager](manager/manager.go) package handles the job
management and stats, and the [process](process/process.go) package handles
the process execution and logging goroutines.

## Usage

Put the apcron binary inside your main app container an run it from your
entrypoint. We use it in the same prod container, specifying different CMD.

Example for our python containers:

The crontab **crons.yaml** file: 

```yaml
# crontab example. Unlike original crontab format, comments allowed ;) 

send-mail:
    second: "10"
    minute: "*/5"
    command: python manage.py send_mail

another-silly-example:  #  This is only a label. 
    second: "0" 
    minute: '10,30,50'   
    command: echo "Hello this is stdout"; sleep 1; echo "Hello stderr" >2 ; exit 0 
    shell: sh
```

The **Dockerfile**. Includes apcron executable along with your main app container:

```dockerfile
[...]

COPY apcron /usr/local/bin/apcron
COPY src/crons.yml /etc/crons.yml


# app src code
WORKDIR /app 
COPY src  ./

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["run-uwsgi"]
EXPOSE 8080 1717

[...]
```

The **entrypoint.sh**: 

```sh
#!/bin/bash

set -e

case $1 in
    run-uwsgi)
        exec uwsgi --ini=/etc/uwsgi/uwsgi.ini
        ;;

    run-crons)
        echo "→ Starting cron"
        exec apcron -v -cmd-prefix "python manage.py" -file /etc/crons.yml
        ;;
    *)
        exec "$@"
        ;;
esac
```

So when launched with default CMD, the container runs your default app. When
launched with *run-crons* CMD, it will run a container dedicated to your cron
jobs.

## History

It was 2016 and apsl.net was running production apps at scale with
kubernetes. We needed a simple solution to run legacy crons inside our
containers. So we forked
[anarcher/go-cron](https://github.com/anarcher/go-cron), and added
[sentry.io](https://sentry.io) support and a yaml crontab format.

By the end of 2017 we needed better logging and debug features, and changed
most of the original process exec code in order to add realtime stdout and
stderr logging (before process finishes) and introduce some job stats.

At 2019 name was changed to apcron. A big refactor was introduced, cleaning
code, lowering memory footprint, fixing bugs and introducing options to
exetute subprocess with no shell by default. Crontab file format deleted.
Code is migrated to go mod and semver versioning.

## License

This project is licensed under the GPL License - see the [LICENSE](LICENSE) file for details.