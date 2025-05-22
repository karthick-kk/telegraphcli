# telegraphcli

A command-line interface for interacting with [telegra.ph](https://telegra.ph/) publishing platform.

## Features

- User management (create, edit, view, revoke)
- Page management (create, list, get, edit, delete, views)
- Markdown support for creating and editing pages
- Robust error handling with automatic retries
- Verbose mode for debugging

## Installation

### From source

```bash
# Clone the repository
git clone https://github.com/karthick-kk/telegraphcli.git
cd telegraphcli

# Build the application
go build
```

## Usage

### User Management

Create a new user:

```bash
./telegraphcli user create
```

View current user information:

```bash
./telegraphcli user view
```

Edit current user information:

```bash
./telegraphcli user edit
```

Revoke access token and generate a new one:

```bash
./telegraphcli user revoke
```

### Page Management

Create a new page from a Markdown file:

```bash
./telegraphcli page create example.md "My Telegraph Post"
```

List your pages:

```bash
./telegraphcli page list
```

Get a page by path:

```bash
./telegraphcli page get my-telegraph-post-05-22
```

Edit a page with a Markdown file:

```bash
./telegraphcli page edit my-telegraph-post-05-22 updated-post.md
```

Delete a page:

```bash
./telegraphcli page delete my-telegraph-post-05-22
```

Get page view count:

```bash
./telegraphcli page views my-telegraph-post-05-22
```

### Using the Wrapper Script

For convenience, a wrapper script is provided:

```bash
./telegraph.sh create-user
./telegraph.sh create-post example.md "My Telegraph Post"
./telegraph.sh list-posts
./telegraph.sh view-post my-telegraph-post-05-22
./telegraph.sh edit-post my-telegraph-post-05-22 updated-post.md
./telegraph.sh delete-post my-telegraph-post-05-22
./telegraph.sh get-views my-telegraph-post-05-22
```

## Troubleshooting

If you encounter connection issues with the Telegraph API, try the following:

1. Use the `--verbose` flag to see detailed error information:
   ```bash
   ./telegraphcli page create example.md "My Post" --verbose
   ```

2. Connection reset errors may occur due to:
   - Temporary Telegraph API issues
   - Network connectivity problems
   - Rate limiting (the client now implements automatic retries)

3. Check your access token validity using:
   ```bash
   ./telegraphcli user view
   ```

4. If problems persist, try creating a new user account:
   ```bash
   ./telegraphcl user revoke
   ```

## License

MIT License
