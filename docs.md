# API Docs

API docs for `goto`. Made from [docs.go](docs.go) and more.

## Authentication

For endpoints that require authentication, you must provide the following:

-   **Username:** Sent as a form body parameter named `username`.
-   **Token:** Sent in the `Authentication` header as a Bearer token (e.g., `Authentication: your_access_token`). ~~Not quite to HTTP spec but I don't really care.~~

You receive an access token upon successful registration using the `/api/join` endpoint.

## Data Structures

### Page

Represents an API endpoint.

| Field         | Type    | Description                                  | Default |
| ------------- | ------- | -------------------------------------------- | ------- |
| `Endpoint`    | string  | The URL path of the endpoint.                |         |
| `Requires_auth` | boolean | Whether the endpoint requires authentication. |         |
| `Method`      | string  | The HTTP method (e.g., GET, POST).          | "GET"   |
| `Description` | string  | A brief description of the endpoint.         |         |
| `Params`      | \[Param] | Array of parameters accepted in URL query. |         |
| `Headers`     | \[Param] | Array of headers accepted. |         |
| `Body`        | \[Param] | Array of parameters accepted in the request body. |         |

### Param

Represents a parameter for an endpoint.

| Field         | Type    | Description                                              | Default  |
| ------------- | ------- | -------------------------------------------------------- | -------- |
| `Name`        | string  | The name of the parameter.                               |          |
| `Description` | string  | A description of the parameter.                          |          |
| `Required`    | boolean | Whether the parameter is required.                      | `false`  |
| `Type`        | string  | The data type of the parameter (e.g., string, boolean). | "string" |

## Endpoints

### `/api`

-   **Method:** `GET`
-   **Requires Authentication:** `false`
-   **Description:** General health check and information endpoint.
-   **Parameters:** None
-   **Headers:** None
-   **Body:** None

### `/api/docs`

-   **Method:** `GET`
-   **Requires Authentication:** `false`
-   **Description:** Documentation - what you're viewing right now!
-   **Parameters:** None
-   **Headers:** None
-   **Body:** None

### `/api/join`

-   **Method:** `POST`
-   **Requires Authentication:** `false`
-   **Description:** Register an account - returns an access token.
-   **Parameters:** None
-   **Headers:** None
-   **Body:**
    -   `username` (string, required): Your chosen username.

### `/api/create`

-   **Method:** `POST`
-   **Requires Authentication:** `true`
-   **Description:** Create a short link.
-   **Parameters:** None
-   **Headers:** None
-   **Body:**
    -   `name` (string, optional): Name of your short link.
    -   `link` (string, required): The long link you want to shorten - is random if not specified.
    -   `private` (boolean, optional): Whether or not authentication must be provided to view statistics from api.

### `/api/link/:name`

-   **Method:** `GET`
-   **Requires Authentication:** `false` (unless the link is private)
-   **Description:** Get statistics of a link - requires authentication if link is private.
-   **Parameters:**
    -   `name` (string, required): Name of the link.
-   **Headers:** None
-   **Body:** None

### `/api/links/:username`

-   **Method:** `GET`
-   **Requires Authentication:** `false` (unless requesting private links)
-   **Description:** Get all the public links of a user - requires authentication for private links.
-   **Parameters:**
    -   `username` (string, required): Username of links to look up.
-   **Headers:** None
-   **Body:** None