const http = require('http');
const os = require('os')
const fs = require('fs')

console.log('user', os.userInfo({ encoding: 'utf8' }))

const port = process.env.PORT || 8080;
const hostname = '0.0.0.0';

const server = http.createServer((req, res) => {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Hello World');
});

server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}/`);
});