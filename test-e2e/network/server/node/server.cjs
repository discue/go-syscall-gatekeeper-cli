const http = require('http');
const os = require('os');

const port = process.env.PORT || 8080;
const hostname = '0.0.0.0'; // Use for local server on the machine

const server = http.createServer((req, res) => {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Hello World');
});

// Example of signal handling using "process" API:
process.on('SIGINT', () => {
    console.log('Received SIGINT signal, closing server.');
    server.close(() => {
        console.log('Server closed successfully!');
        process.exit(0); // Exit the process after successful closure
    });
});

// Additional Error Handling (for graceful shutdown)
process.once('SIGTERM', () => {
    console.log("Received SIGTERM signal, shutting down server gracefully.");
    server.close(() => {
        console.log("Server closed successfully!");
        process.exit(0); // Exit the process after successful closure
    });
});

server.listen(port, hostname, () => {
    console.log(`Server running at http://${hostname}:${port}/`);
});
