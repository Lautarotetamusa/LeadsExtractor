package onedrive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/manuvariego/golang-onedrive/onedrive"
)

type Node struct {
    Path            string
    downloadUrl     string
    IsFolder        bool

    parent  *Node
    childs  []*Node
}

const baseUrl = "https://graph.microsoft.com/v1.0/me"

func BuildTree(od *onedrive.OneDriveClient, root *Node) {
    childs, err := od.GetFiles()
    if err != nil {
        fmt.Printf("error getting files, pwd: %s", root.Path)
    }

    for _, child := range childs {
        childNode := &Node{
            path: root.path + "/" + child.Name,
            downloadUrl: child.DownloadUrl,
            isFolder: child.IsFolder != nil,

            parent: root,
        }
        if childNode.isFolder {
            BuildTree(od, childNode)
        }
    }
}

func getFiles(client *http.Client, path string) ([]Node, error) {
	url := baseUrl + path + ":/children"

	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, err
	}

	var response struct {
		Value []Node `json:"value"`
	}

    if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
        return nil, err
    }

	return response.Value, nil
}

func main() {
    scopes := []string{"User.Read", "offline_access", "Sites.Read.All", "Files.ReadWrite.All"}
	tenantID := os.Getenv("MS_TENANT_ID")
	appID := os.Getenv("MS_OPENGRAPH_APP_ID")
	clientSecret := os.Getenv("MS_OPENGRAPH_CLIENT_SECRET")
	sharePoint := os.Getenv("SHAREPOINT_PATH")
	oauthconf := onedrive.NewOauthConfig(tenantID, appID, clientSecret, scopes)

	//Checks if token.json exists, if it doesn't it is created with a new code from user
	if !onedrive.CheckTokenFile() {
		token := onedrive.GetInitialTokens(oauthconf)
        err := onedrive.SaveToken(token)
		if err != nil {
			panic("Error saving token")
		}
	}

	client, _ := onedrive.GetClient(oauthconf)
	rootUrl := onedrive.GetRootUrl(sharePoint)

	//Creates onedriveclient with the data
	od := onedrive.OneDriveClient{Client: client, Path: onedrive.Path{CurrentPath: rootUrl}}
}
