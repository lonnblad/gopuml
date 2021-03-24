[![Build Status](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Factions-badge.atrox.dev%2Flonnblad%2Fgopuml%2Fbadge%3Fref%3Dmain&style=flat)](https://actions-badge.atrox.dev/lonnblad/gopuml/goto?ref=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/lonnblad/gopuml)](https://goreportcard.com/report/github.com/lonnblad/gopuml)
[![Coverage Status](https://coveralls.io/repos/github/lonnblad/gopuml/badge.svg?branch=main)](https://coveralls.io/github/lonnblad/gopuml?branch=main)

# gopuml

This a tool to compile [Plant UML](https://plantuml.com/) into files and links.

## Table of Contents

- [Usage](#usage)
- [Examples](#examples)

## Usage

### Install

> go get -u github.com/lonnblad/gopuml

### Compiling UML

The command used to compile the Plant UML to different formats.

> gopuml build

#### Options

- **-f, --format**

  The format to use when compiling the Plant UML, defaults to: `svg`.

  Supported formatters are:

  - `png`, will format the content as .png
  - `svg`, will format the content as .svg
  - `txt`, will format the content as .txt

- **--server**

  The Server URL to use when the style used is `link`, defaults to: `https://www.planttext.com/api/plantuml`.

- **--style**

  The style to use when compiling the Plant UML, defaults to: `file`.

  Supported styles are:

  - `file`, will write the formatted content to a file
  - `link`, will write a link to the formatted content to stdout
  - `out`, will write the formatted content to stdout

## Examples

These examples can be found [here](example).

### [example.puml](example/example.puml)

The source Plant UML.

```puml
@startuml Example
Bob -> Alice : hello
@enduml
```

### Compile files

#### Compiles [example.png](example/example.png).

> gopuml build -f png example/example.puml

![example.png](example/example.png)

#### Compiles [example.svg](example/example.svg).

> gopuml build -f svg example/example.puml

![example.svg](example/example.svg)

#### Compiles [example.txt](example/example.txt).

> gopuml build -f txt example/example.puml

```txt
     ┌───┐          ┌─────┐
     │Bob│          │Alice│
     └─┬─┘          └──┬──┘
       │    hello      │
       │──────────────>│
     ┌─┴─┐          ┌──┴──┐
     │Bob│          │Alice│
     └───┘          └─────┘
```

### Generate links

#### Generates a link for the [example.png](https://www.planttext.com/api/plantuml/png/SYWkIImgAStDKN2jICmjo4dbSifFKj2rKt3CoKnELR1Io4ZDoSddSaZDIodDpG44003__m00).

> gopuml build -f png --style link example/example.puml

#### Generates a link for the [example.svg](https://www.planttext.com/api/plantuml/svg/SYWkIImgAStDKN2jICmjo4dbSifFKj2rKt3CoKnELR1Io4ZDoSddSaZDIodDpG44003__m00).

> gopuml build -f svg --style link example/example.puml

#### Generates a link for the [example.txt](https://www.planttext.com/api/plantuml/txt/SYWkIImgAStDKN2jICmjo4dbSifFKj2rKt3CoKnELR1Io4ZDoSddSaZDIodDpG44003__m00).

> gopuml build -f txt --style link example/example.puml
