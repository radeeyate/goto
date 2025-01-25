package main

type Page struct {
	Endpoint      string  `json:"endpoint"` // url path of the endpoint
	Requires_auth bool    `json:"requires_auth"` // whether the endpoint needs authentication
	Method        string  `json:"method"        default:"GET"` // http method of endpoint
	Description   string  `json:"description"` // brief description of the endpoint
	Params        []Param `json:"params"` // an array of params in a URL. not to be confused with query params
	Headers       []Param `json:"headers"` // array of headers accepted
	Body          []Param `json:"body"` // array of parameters accepted in the request body
}

type Param struct {
	Name        string `json:"name"` // name of parameter/header/form value
	Description string `json:"description"` // description of parameter/header/form value
	Required    bool   `json:"required"    default:"false"` // whether the parameter is required
	Type        string `json:"type"        default:"string"` // type of parameter/header/form value
}

var pages = []Page{
	{
		Endpoint:      "/api",
		Requires_auth: false,
		Method:        "GET",
		Description:   "general health check and information endpoint",
		Params:        []Param{},
		Headers:       []Param{},
		Body:          []Param{},
	},
	{
		Endpoint:      "/api/docs",
		Requires_auth: false,
		Method:        "GET",
		Description:   "documentation - what you're viewing right now!",
		Params:        []Param{},
		Headers:       []Param{},
		Body:          []Param{},
	},
	{
		Endpoint:      "/api/join",
		Requires_auth: false,
		Method:        "POST",
		Description:   "register an account - returns an access token",
		Params:        []Param{},
		Headers:       []Param{},
		Body: []Param{
			{
				Name:        "username",
				Description: "your chosen username",
				Required:    true,
				Type:        "string",
			},
		},
	},
	{
		Endpoint:      "/api/create",
		Requires_auth: true,
		Method:        "POST",
		Description:   "create a short link",
		Params:        []Param{},
		Headers:       []Param{},
		Body: []Param{
			{
				Name:        "name",
				Description: "name of your short link",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "link",
				Description: "the long link you want to shorten - is random if not specified",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "private",
				Description: "whether or not authentication must be provided to view statistics from api",
				Required:    false,
				Type:        "boolean",
			},
		},
	},
	{
		Endpoint:      "/api/link/:name",
		Requires_auth: false,
		Method:        "GET",
		Description:   "get statistics of a link - requires authentication if link is private",
		Headers:       []Param{},
		Body:          []Param{},
		Params: []Param{
			{
				Name:        "name",
				Description: "name of the link",
				Required:    true,
				Type:        "string",
			},
		},
	},
	{
		Endpoint:      "/api/links/:username",
		Requires_auth: false,
		Method:        "GET",
		Description:   "get all the public links of a user - requires authentication for private links",
		Headers:       []Param{},
		Body:          []Param{},
		Params: []Param{
			{
				Name:        "username",
				Description: "username of links to look up",
				Required:    true,
				Type:        "string",
			},
		},
	},
}
