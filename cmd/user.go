package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	telegraph "source.toby3d.me/toby3d/telegraph/v2"

	"telegraphcli/pkg/token"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Operations related to Telegraph account management",
	Long:  `With this command, one can handle an user account to manage your pages.`,
}

// userCreateCmd represents the user create command
var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an user",
	Long: `Create a new Telegraph user.
A token is generated and stored at ~/.telegraphcl/telegraph.token`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// client := http.DefaultClient // Not used directly anymore
		verbose, _ := cmd.Flags().GetBool("verbose")

		var shortNameInput, authorNameInput string
		fmt.Print("Enter short name: ")
		fmt.Scanln(&shortNameInput)
		fmt.Print("Enter author name: ")
		fmt.Scanln(&authorNameInput)

		shortName, err := telegraph.NewShortName(shortNameInput)
		if err != nil {
			cmd.PrintErrf("Failed to create short name: %v\n", err)
			return
		}

		authorName, err := telegraph.NewAuthorName(authorNameInput)
		if err != nil {
			cmd.PrintErrf("Failed to create author name: %v\n", err)
			return
		}

		createAccount := telegraph.CreateAccount{
			ShortName:  *shortName,
			AuthorName: authorName,
		}

		var account *telegraph.Account
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			account, e = createAccount.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during create account: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to create account after retries: %v\\n", err)
			return
		}

		if err := token.SaveToken(account.AccessToken); err != nil {
			cmd.PrintErrf("Failed to save token: %v\n", err)
			return
		}

		cmd.Println("Account created successfully!")
		cmd.Println("Short Name:", account.ShortName)
		cmd.Println("Author Name:", account.AuthorName)
		cmd.Println("Access Token:", account.AccessToken)
	},
}

// userEditCmd represents the user edit command
var userEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit current user information",
	Long:  `Edit current user information such as short name and author name.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// client := http.DefaultClient // Not used directly anymore
		verbose, _ := cmd.Flags().GetBool("verbose")

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\n", err)
			return
		}

		// Get current info first
		// Create our own AccountField values since the API has changed
		fieldShortName := telegraph.FieldShortName
		fieldAuthorName := telegraph.FieldAuthorName
		getAccountInfo := telegraph.GetAccountInfo{
			AccessToken: accessToken,
			Fields:      []telegraph.AccountField{fieldShortName, fieldAuthorName},
		}
		var currentAccount *telegraph.Account
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			currentAccount, e = getAccountInfo.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during get account info for edit: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to get account info after retries: %v\\n", err)
			return
		}

		cmd.Println("Current Short Name:", currentAccount.ShortName)
		cmd.Println("Current Author Name:", currentAccount.AuthorName.String())

		var shortNameInput, authorNameInput string
		fmt.Print("Enter new short name (leave blank to keep current): ")
		fmt.Scanln(&shortNameInput)
		fmt.Print("Enter new author name (leave blank to keep current): ")
		fmt.Scanln(&authorNameInput)

		editAccount := telegraph.EditAccountInfo{
			AccessToken: accessToken,
		}

		var newShortName *telegraph.ShortName // Type is *telegraph.ShortName
		var newAuthorName *telegraph.AuthorName // Type is *telegraph.AuthorName

		if shortNameInput != "" {
			shortNameVal, err := telegraph.NewShortName(shortNameInput)
			if err != nil {
				cmd.PrintErrf("Failed to create short name: %v\\n", err)
				return
			}
			newShortName = shortNameVal // Assign the pointer
			editAccount.ShortName = newShortName
		}

		if authorNameInput != "" {
			authorNameVal, err := telegraph.NewAuthorName(authorNameInput)
			if err != nil {
				cmd.PrintErrf("Failed to create author name: %v\\n", err)
				return
			}
			newAuthorName = authorNameVal // Assign the pointer
			editAccount.AuthorName = newAuthorName
		}

		var updatedAccount *telegraph.Account
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			updatedAccount, e = editAccount.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during edit account info: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to edit account info after retries: %v\\n", err)
			return
		}

		cmd.Println("Account updated successfully!")
		cmd.Println("Short Name:", updatedAccount.ShortName)
		cmd.Println("Author Name:", updatedAccount.AuthorName)
	},
}

// userRevokeCmd represents the user revoke command
var userRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke and regenerate access token",
	Long:  `Revoke the current access token and generate a new one.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		verbose, _ := cmd.Flags().GetBool("verbose")

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\\n", err)
			return
		}

		// Revoke access token
		revokeAccessToken := telegraph.RevokeAccessToken{
			AccessToken: accessToken,
		}
		var newAccount *telegraph.Account
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			newAccount, e = revokeAccessToken.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during revoke access token: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to revoke access token after retries: %v\\n", err)
			return
		}

		if err := token.SaveToken(newAccount.AccessToken); err != nil {
			cmd.PrintErrf("Failed to save new token: %v\\n", err)
			return
		}

		cmd.Println("Access token revoked and new token generated successfully!")
		cmd.Println("New Access Token:", newAccount.AccessToken)
		cmd.Println("New Auth URL:", newAccount.AuthURL)
	},
}

// userViewCmd represents the user view command
var userViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View current user information",
	Long:  `View current user information such as short name, author name, and page count.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		verbose, _ := cmd.Flags().GetBool("verbose")

		accessToken, err := token.GetToken()
		if err != nil {
			cmd.PrintErrf("Failed to get token: %v\\n", err)
			return
		}

		fieldShortName := telegraph.FieldShortName
		fieldAuthorName := telegraph.FieldAuthorName
		fieldPageCount := telegraph.FieldPageCount
		getAccountInfo := telegraph.GetAccountInfo{
			AccessToken: accessToken,
			Fields:      []telegraph.AccountField{fieldShortName, fieldAuthorName, fieldPageCount},
		}
		var account *telegraph.Account
		err = retry(func() error {
			clientWithHeaders := &http.Client{
				Timeout: httpClient.Timeout,
				Transport: &customTransport{
					base:      httpClient.Transport,
					userAgent: userAgent,
				},
			}
			var e error
			account, e = getAccountInfo.Do(ctx, clientWithHeaders)
			if e != nil && verbose {
				cmd.Printf("Request failed during get account info for view: %v\\n", e)
			}
			return e
		}, 3)

		if err != nil {
			cmd.PrintErrf("Failed to get account info after retries: %v\\n", err)
			return
		}

		cmd.Println("Account Information:")
		cmd.Println("Short Name:", account.ShortName)
		cmd.Println("Author Name:", account.AuthorName.String())
		cmd.Println("Page Count:", account.PageCount)
		cmd.Println("Auth URL:", account.AuthURL)
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userEditCmd)
	userCmd.AddCommand(userRevokeCmd)
	userCmd.AddCommand(userViewCmd) 
}
