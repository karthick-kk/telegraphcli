package cmd

import (
	"context"
	"net/http"
	"net/url" // Added import
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spf13/cobra"
	telegraph "source.toby3d.me/toby3d/telegraph/v2"
	"golang.org/x/net/html/atom"

	"telegraphcli/pkg/markdown"
	"telegraphcli/pkg/token"
)

// pageCmd represents the page command
var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Manage your Telegra.ph pages",
	Long: `With this option, one can manage pages.
Markdown files are used to create and edit pages.`,
}

// pageCreateCmd represents the page create command
var pageCreateCmd = &cobra.Command{
	Use:   "create <markdown-path> <title>",
	Short: "Create Page from a Markdown file",
	Args:  cobra.ExactArgs(2),
	Long:  `Create a new Telegra.ph page from a Markdown file.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second) // Increased timeout for multiple API calls
		defer cancel()

		verbose, _ := cmd.Flags().GetBool("verbose")
		markdownPath := args[0]
		title := args[1]

		if verbose {
			cmd.Println("Parsing markdown file:", markdownPath)
		}

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\\n", err)
			return
		}

		// Get Author Name and URL from account info
		var authorName, authorURL string
		getAccountInfo := telegraph.GetAccountInfo{
			AccessToken: accessToken,
			Fields:      []telegraph.AccountField{telegraph.FieldAuthorName, telegraph.FieldAuthorURL},
		}
		var accountInfo *telegraph.Account
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout:   httpClient.Timeout,
				Transport: &customTransport{base: httpClient.Transport, userAgent: userAgent},
			}
			var e error
			accountInfo, e = getAccountInfo.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during get account info for page creation: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to get account info for author details: %v. Page will be created without author info.\\\\n", err)
			// Continue without author info if fetching fails
		} else {
			if accountInfo != nil {
				// Assuming AuthorName and AuthorURL are struct types based on previous build errors
				// (value types not comparable to nil, but have .String() methods).
				// If their .String() methods panic on zero values, it's an issue in the
				// library or the underlying net/url call in this environment.

				// Handle AuthorName
				// If accountInfo.AuthorName.String() also panics for zero values, a similar issue exists.
				nameStr := accountInfo.AuthorName.String()
				if nameStr != "" {
					authorName = nameStr
				}

				// Handle AuthorURL
				// The panic occurs if accountInfo.AuthorURL.String() is called when its internal *net/url.URL is nil.
				// We assume the telegraph.URL struct has an exported field 'URL' of type *net/url.URL.
				if accountInfo.AuthorURL.URL != nil { // Check the internal *net/url.URL pointer
					urlStr := accountInfo.AuthorURL.String()
					if urlStr != "" {
						authorURL = urlStr
					}
				} else {
					// authorURL remains empty. Log if verbose.
					if verbose {
						cmd.Println("AuthorURL from account info has a nil internal net/url.URL; cannot convert to string.")
					}
				}

				if verbose {
					cmd.Printf("Fetched author details: Name='%s', URL='%s'\\\\n", authorName, authorURL)
				}
			}
		}

		// Parse markdown file
		nodes, err := markdown.Parse(markdownPath)
		if err != nil {
			cmd.PrintErrf("Failed to parse markdown: %v\n", err)
			return
		}

		if verbose {
			cmd.Printf("Successfully parsed markdown file, found %d nodes\n", len(nodes))
		}

		// Create page
		pageTitle, err := telegraph.NewTitle(title)
		if err != nil {
			cmd.PrintErrf("Failed to create title: %v\\n", err)
			return
		}
		
		if verbose {
			cmd.Println("Creating page with title:", title)
			cmd.Println("Using access token:", accessToken[:10]+"...")
		}
		
		// Prepare AuthorName and AuthorURL for CreatePage struct
		var telegraphAuthorName *telegraph.AuthorName
		if authorName != "" {
			newName, err := telegraph.NewAuthorName(authorName)
			if err != nil {
				cmd.PrintErrf("Invalid author name format '%s': %v. Proceeding without author name.\\\\\\\\n", authorName, err)
				// telegraphAuthorName remains nil
			} else {
				telegraphAuthorName = newName
			}
		}

		var telegraphAuthorURL *telegraph.URL
		if authorURL != "" {
			parsedBaseURL, parseErr := url.Parse(authorURL) // Parse string to net/url.URL
			if parseErr != nil {
				cmd.PrintErrf("Invalid author URL string format '%s': %v. Proceeding without author URL.\\\\\\\\n", authorURL, parseErr)
				// telegraphAuthorURL remains nil
			} else {
				// telegraph.NewURL expects *url.URL and returns *telegraph.URL (no error)
				newTelegraphURL := telegraph.NewURL(parsedBaseURL) 
				telegraphAuthorURL = newTelegraphURL
			}
		}
		
		createPage := telegraph.CreatePage{
			AccessToken: accessToken,
			Title:       *pageTitle,
			Content:     nodes,
			AuthorName:  telegraphAuthorName, // Use pointer to telegraph.AuthorName
			AuthorURL:   telegraphAuthorURL,  // Use pointer to telegraph.URL
		}
		
		// Create a request using our custom HTTP client with user agent
		var page *telegraph.Page
		err = retry(func() error {
			if verbose {
				cmd.Println("Sending request to Telegraph API...")
			}
			
			// Create a custom HTTP client that adds user agent header to each request
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base: httpClient.Transport,
					userAgent: userAgent,
				},
			}
			
			var err error
			page, err = createPage.Do(ctx, clientWithHeaders)
			if err != nil && verbose {
				cmd.Printf("Request failed: %v\n", err)
			}
			return err
		}, 3)
		
		if err != nil {
			cmd.PrintErrf("Failed to create page after retries: %v\n", err)
			return
		}

		cmd.Println("Page created successfully!")
		cmd.Println("Title:", page.Title)
		cmd.Println("URL:", page.URL)
		cmd.Println("Path:", page.Path)
	},
}

// pageListCmd represents the page list command
var pageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your Telegra.ph pages",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// client := http.DefaultClient // Not used directly anymore

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\\n", err)
			return
		}

		// Get page list
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		getPageList := telegraph.GetPageList{
			AccessToken: accessToken,
			Limit:       uint16(limit),
			Offset:      uint(offset),
		}
		
		var pageList *telegraph.PageList
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			pageList, e = getPageList.Do(ctx, clientWithHeaders)
			if e != nil {
				if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
					cmd.Printf("Request failed during page list: %v\\n", e)
				}
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to get page list after retries: %v\\n", err)
			return
		}

		cmd.Printf("Total pages: %d\\n", pageList.TotalCount)
		cmd.Println("Pages:")
		for i, page := range pageList.Pages {
			cmd.Printf("%d. %s (%s)\n", i+1, page.Title.String(), page.URL.String())
		}
	},
}

// pageGetCmd represents the page get command
var pageGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get page with Telegra.ph path",
	Args:  cobra.ExactArgs(1),
	Long:  `Get details of a page by its path.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// client := http.DefaultClient // Not used directly anymore

		path := args[0]

		// Get page info
		getPage := telegraph.GetPage{
			Path:          path,
			ReturnContent: true,
		}

		var page *telegraph.Page
		err := retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			page, e = getPage.Do(ctx, clientWithHeaders)
			if e != nil {
				if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
					cmd.Printf("Request failed during get page: %v\\n", e)
				}
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to get page after retries: %v\\n", err)
			return
		}

		cmd.Println("Title:", page.Title)
		cmd.Println("Author:", page.AuthorName)
		cmd.Println("URL:", page.URL)
		cmd.Println("Views:", page.Views)
		// The API doesn't return creation time anymore
	},
}

// pageDeleteCmd represents the page delete command
var pageDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a Telegra.ph page by editing its content to be empty",
	Args:  cobra.ExactArgs(1),
	Long:  `Effectively deletes a Telegra.ph page by clearing its title, content, author name, and author URL.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		verbose, _ := cmd.Flags().GetBool("verbose")
		path := args[0]

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\\n", err)
			return
		}

		if verbose {
			cmd.Printf("Attempting to 'delete' page with path: %s\\n", path)
		}

		// Prepare new minimal title
		deletedTitle, err := telegraph.NewTitle("Deleted") // Or use a single space if API allows and preferred
		if err != nil {
			cmd.PrintErrf("Failed to create 'Deleted' title: %v\\n", err)
			return
		}

		// Prepare minimal content to satisfy API requirements
		pTag, _ := telegraph.NewTag(atom.P)
		pElem := telegraph.NewNodeElement(pTag)
		textNode := telegraph.Node{
			Text: "[deleted]",
		}
		pElem.Children = append(pElem.Children, textNode)
		deletedContent := []telegraph.Node{
			telegraph.Node{
				Element: pElem,
			},
		}

		editPage := telegraph.EditPage{
			AccessToken:   accessToken,
			Path:          path,
			Title:         *deletedTitle,
			Content:       deletedContent, // Use minimal non-empty content
			AuthorName:    nil,            // Clear author name
			AuthorURL:     nil,            // Clear author URL
			ReturnContent: false,
		}

		err = retry(func() error {
			if verbose {
				cmd.Println("Sending request to Telegraph API to 'delete' page...")
			}
			clientWithHeaders := &http.Client{
				Timeout:   httpClient.Timeout,
				Transport: &customTransport{base: httpClient.Transport, userAgent: userAgent},
			}
			_, e := editPage.Do(ctx, clientWithHeaders) // We don't need the returned page
			if e != nil && verbose {
				cmd.Printf("Request failed: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to 'delete' page at path '%s' after retries: %v\\n", path, err)
			return
		}

		cmd.Printf("Page at path '%s' has been 'deleted' (content cleared).\\n", path)
	},
}

// pageEditCmd represents the page edit command
var pageEditCmd = &cobra.Command{
	Use:   "edit <path> <markdown-path>",
	Short: "Edit page with Telegra.ph path",
	Args:  cobra.ExactArgs(2),
	Long:  `Edit an existing Telegra.ph page with a Markdown file.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// client := http.DefaultClient // Not used directly anymore

		path := args[0]
		markdownPath := args[1]
		verbose, _ := cmd.Flags().GetBool("verbose")

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\\n", err)
			return
		}

		// Parse markdown file
		nodes, err := markdown.Parse(markdownPath)
		if err != nil {
			cmd.PrintErrf("Failed to parse markdown: %v\\n", err)
			return
		}

		// Get current page to keep the title
		getPage := telegraph.GetPage{
			Path:          path,
			ReturnContent: false,
		}
		var currentPage *telegraph.Page
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			currentPage, e = getPage.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during get current page for edit: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to get current page after retries: %v\\n", err)
			return
		}

		// Check if a new title was provided
		var pageTitle telegraph.Title
		if currentPage.Title != nil {
			pageTitle = *currentPage.Title
		}
		
		newTitle, _ := cmd.Flags().GetString("title")
		if newTitle != "" {
			newPageTitle, err := telegraph.NewTitle(newTitle)
			if err != nil {
				cmd.PrintErrf("Failed to create title: %v\n", err)
				return
			}
			pageTitle = *newPageTitle
		}

		// Edit page
		editPage := telegraph.EditPage{
			AccessToken: accessToken,
			Path:        path,
			Title:       pageTitle,
			Content:     nodes,
		}

		var page *telegraph.Page
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			page, e = editPage.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during edit page: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to edit page after retries: %v\\n", err)
			return
		}

		cmd.Println("Page edited successfully!")
		cmd.Println("Title:", page.Title)
		cmd.Println("URL:", page.URL)
		cmd.Println("Path:", page.Path)
	},
}

// pageViewsCmd represents the page views command
var pageViewsCmd = &cobra.Command{
	Use:   "views <path>",
	Short: "Count views on your Telegra.ph page",
	Args:  cobra.ExactArgs(1),
	Long:  `Get the count of views on a particular page.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// client := http.DefaultClient // Not used directly anymore

		path := args[0]
		
		// Parse year, month, day, and hour if provided
		year, _ := cmd.Flags().GetInt("year")
		month, _ := cmd.Flags().GetInt("month")
		day, _ := cmd.Flags().GetInt("day")
		hour, _ := cmd.Flags().GetInt("hour")

		// Get page views
		getViews := telegraph.GetViews{
			Path:  path,
			Year:  uint16(year),
			Month: uint8(month),
			Day:   uint8(day),
			Hour:  uint8(hour), // Corrected type for Hour back to uint8
		}

		var views *telegraph.PageViews // Corrected type to telegraph.PageViews
		err := retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			views, e = getViews.Do(ctx, clientWithHeaders)
			if e != nil {
				if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
					cmd.Printf("Request failed during get views: %v\\n", e)
				}
			}
			return e
		}, 3)
		
		if err != nil {
			cmd.PrintErrf("Failed to get views after retries: %v\\n", err)
			return
		}

		cmd.Println("Views:", views.Views)
	},
}

// retry attempts a function with retries using exponential backoff
func retry(fn func() error, attempts int) error { 
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 1 * time.Minute // Example: Max time to keep retrying

	operation := func() error {
		return fn()
	}

	return backoff.Retry(operation, bo)
}

func init() {
	rootCmd.AddCommand(pageCmd)
	pageCmd.AddCommand(pageCreateCmd)
	pageCmd.AddCommand(pageListCmd)
	pageCmd.AddCommand(pageGetCmd)
	pageCmd.AddCommand(pageEditCmd)
	pageCmd.AddCommand(pageDeleteCmd) // Added pageDeleteCmd
	pageCmd.AddCommand(pageViewsCmd)

	// Add global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output for debugging")
	
	// Add flags to commands
	pageListCmd.Flags().IntP("limit", "l", 10, "Limit the number of pages returned")
	pageListCmd.Flags().IntP("offset", "o", 0, "Offset in the list of pages")
	
	pageEditCmd.Flags().StringP("title", "t", "", "New title for the page")
	
	pageViewsCmd.Flags().IntP("year", "y", 0, "Year to filter views")
	pageViewsCmd.Flags().IntP("month", "m", 0, "Month to filter views")
	pageViewsCmd.Flags().IntP("day", "d", 0, "Day to filter views")
	pageViewsCmd.Flags().IntP("hour", "H", 0, "Hour to filter views")
}

// customTransport adds custom headers to HTTP requests
type customTransport struct {
	base      http.RoundTripper
	userAgent string
}

// RoundTrip implements the http.RoundTripper interface
func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Set user agent header
	req.Header.Set("User-Agent", t.userAgent)
	
	// Use the base transport to perform the request
	return t.base.RoundTrip(req)
}
