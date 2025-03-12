package download

import(

)

type SavedDownloadState struct{
	URL string
	FilePath string
	CHUNK_SIZE int
	CompletedParts []bool
	CurrByte int64
	TotalBytes int64
	PartsCount int
	IsPaused bool
	ContentLength int64
	IncompleteParts []int
}

func (h *DownloadHandler) Export(download *Download) (*SavedDownloadState, error){

}

func (h *DownloadHandler) Import(state *SavedDownloadState) (*Download, error){}

// same as save
func (h *DownloadHandler) Serialize(download *Download) ([]byte, error){}
// same as load
func (h *DownloadHandler) Deserialize(data []byte) (*Download, error){}