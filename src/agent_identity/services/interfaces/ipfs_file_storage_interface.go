package interfaces

// IPFSFileStorageInterface uploads a file's bytes to IPFS and returns its CID.
type IPFSFileStorageInterface interface {
	UploadFile(fileBytes []byte, filename string) (*string, error)
}
