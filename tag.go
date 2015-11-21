package metrik

//Tags are a key-value map of strings representing dimensions over which a metric is broken down. For example, {"Region": "South-east"}, {"Country": ["England", "Wales"]}.
type Tags map[string][]string

//Tag contains tag key metadata.
type Tag struct {
	Name        string `json:"name"`        //Name of the tag group - should be URL friendly.
	Description string `json:"description"` //Longer description for HTTP API users.
}
