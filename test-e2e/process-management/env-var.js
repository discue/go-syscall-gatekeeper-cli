if (parseInt(process.env.GATEKEEPER_PID) > 1) {
    process.exit(0)
} else {
    process.exit(123)
}