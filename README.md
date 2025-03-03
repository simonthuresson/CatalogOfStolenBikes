# Catalog of stolen bikes

## Overview

A simple Go server application deployed with Docker. 

## Requirements

 - Docker compose 

## Architectural Overview
![Image description](Overview.png)

### Description of architecture
The entry point (main) begins by initializing the database. Once initialization is complete, the server's endpoints are registered. The functions associated with each endpoint are fetched from the utils directory, which contains several utility files organized by specific domains. Each utility file uses the initialized database to perform queries. After this setup process is complete, the main file starts the server.

## Quick Start
From root directory

```bash
./start-server.sh
```
or 
```bash
docker compose up -d --build
```
this starts the server and database which is accesible from localhost:8080