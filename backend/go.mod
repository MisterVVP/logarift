module github.com/MisterVVP/logarift/backend

go 1.25

require go.mongodb.org/mongo-driver/v2 v2.3.1

replace go.mongodb.org/mongo-driver/v2 => ./internal/mongodriver
