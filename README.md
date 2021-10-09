# InstaAPI

An Instagram API built for the Appointy internship tech task.

## Tasklist

- Create an User
    - [X] Should be a POST request
    - [X] Use JSON request body
    - [X] URL should be ‘/users'
- Get a user using ID
    - [X] Should be a GET request
    - [X] Id should be in the url parameter
    - [X] URL should be ‘/users/<id here>’
- Create a Post
    - [X] Should be a POST request
    - [X] Use JSON request body
    - [X] URL should be ‘/posts'
- Get a post using ID
    - [X] Should be a GET request
    - [X] Id should be in the url parameter
    - [X] URL should be ‘/posts/<id here>’
- List all posts of a user
    - [X] Should be a GET request
    - [X] URL should be ‘/posts/users/<Id here>'

- Quality of Code
    - [X] Reusability
    - [X] Consistency in naming variables, methods, functions, types
    - [X] diomatic i.e. in Go’s style
- [X] Passwords should be securely stored such they can't be reverse engineered
- [X] Make the server thread safe
- [X] Add pagination to the list endpoint
- [ ] Add unit tests