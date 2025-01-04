def read_run_sh():
    """Reads the contents of of /etc/resolve.conf and returns them as a string.
       Returns None if the file doesn't exist or if an error occurs.
       Prints an error message to stderr if the file can't be read.
    """

    try:
        with open("/etc/resolv.conf", "r") as f:
            return f.read()
    except FileNotFoundError:
        print("Error: not found.", file=sys.stderr) # sys needed
        return None
    except Exception as e: # Broad except to catch all other file errors
        print(f"Error reading: {e}", file=sys.stderr)
        return None

if __name__ == "__main__":
    import sys # Added import statement for sys module

    contents = read_run_sh()
    if contents and "nameserver" in contents:
        sys.exit(0)  # Exit with 0 if "nameserver" is found
    else:
        sys.exit(1)  # Exit with 1 if "nameserver" is not found or an error occurred