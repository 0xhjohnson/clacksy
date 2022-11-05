package main

type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")
const pageContextKey = contextKey("page")
