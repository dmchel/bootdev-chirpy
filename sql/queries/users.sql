-- name: CreateUser :one
INSERT INTO public.users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: UpdateUser :one
UPDATE public.users SET email = $1, hashed_password = $2, updated_at = NOW() WHERE id = $3
RETURNING *;

-- name: GetUser :one
SELECT id, created_at, updated_at, email, hashed_password FROM public.users WHERE email = $1;

-- name: DeleteAllUsers :execresult
DELETE FROM public.users;

-- name: UpdateUserToChirpyRed :one
UPDATE public.users SET is_chirpy_red = $1 WHERE id = $1