package metrik

type Tags map[string][]string //Tag is a key-value map of strings representing dimensions over which a metric is broken down. For example, {"Region": "South-east"}, {"Country": ["England", "Wales"]}.

type Tag struct {
	Name        string //Name of the tag group - should be URL friendly.
	Description string //Longer description for users.
}

type TagMetadata struct {
	Name        string `json:"name"`
	Description string `json:"string"`
}

func getTagMetadata(t Tag) TagMetadata {
	return TagMetadata{
		Name:        t.Name,
		Description: t.Description,
	}
}
