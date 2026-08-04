package keys

type ActionShow = string

const (
	ShowBulkDelete      ActionShow = "bulk-delete"
	ShowBulkDeleteForce ActionShow = "bulk-delete-force"

	ShowContainerStop         ActionShow = "container-stop"
	ShowContainerKill         ActionShow = "container-kill"
	ShowContainerRestart      ActionShow = "container-restart"
	ShowContainerRemove       ActionShow = "container-remove"
	ShowContainerCopy         ActionShow = "container-copy"
	ShowContainerExport       ActionShow = "container-export"
	ShowContainerCommitExport ActionShow = "container-commit-export"

	ShowImageRemove ActionShow = "image-remove"
	ShowImageSave   ActionShow = "image-save"

	ShowVolumeRemove ActionShow = "volume-remove"
	ShowVolumePrune  ActionShow = "volume-prune"

	ShowNetworkRemove ActionShow = "network-remove"
	ShowNetworkPrune  ActionShow = "network-prune"

	ShowEventsClear ActionShow = "events-clear"

	ShowBatchStop ActionShow = "batch-stop"
	ShowBatchKill ActionShow = "batch-kill"

	ShowOptionCancel  ActionShow = "cancel"
	ShowOptionConfirm ActionShow = "confirm"
	ShowOptionForce   ActionShow = "force"
)
