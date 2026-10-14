-- Base de datos para autenticación y usuarios
CREATE TABLE IF NOT EXISTS roles (
  id bigint primary key generated always as identity,
  name text not null unique
);

CREATE TABLE IF NOT EXISTS permissions (
  id bigint primary key generated always as identity,
  name text not null unique
);

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id bigint references roles (id),
  permission_id bigint references permissions (id),
  primary key (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS users (
  id bigint primary key generated always as identity,
  nombres text not null,
  paterno text not null,
  materno text,
  rut text unique,
  password text not null,
  email text not null unique,
  picture_url text,
  role_id bigint references roles (id),
  career_code text
);
