package db

import "log"

func Migrate() {
    query := `
    PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT DEFAULT '',
    phone TEXT UNIQUE DEFAULT '',
    designation TEXT DEFAULT '',
    role TEXT CHECK (role IN ('buisness_owner', 'printing', 'plant')),
    profile_url TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS companies (
    company_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    address TEXT,
    logo TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS company_outlets (
    id TEXT PRIMARY KEY,
    company_id TEXT NOT NULL,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS orders (
    order_id TEXT PRIMARY KEY,
    company_id TEXT NOT NULL,
    label_id TEXT NOT NULL,
    variant TEXT NOT NULL,
    qty INTEGER NOT NULL,
    cap_color TEXT NOT NULL,
    volume INTEGER NOT NULL,
    status TEXT CHECK (status IN ('placed', 'printing', 'processing','dispatch')) NOT NULL DEFAULT 'placed',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
    FOREIGN KEY (label_id) REFERENCES labels(label_id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS labels (
    label_id TEXT PRIMARY KEY,
    company_id TEXT NOT NULL,
    name TEXT NOT NULL,
    label_url TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(company_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_companies_user_id ON companies(user_id);
CREATE INDEX IF NOT EXISTS idx_outlets_company_id ON company_outlets(company_id);
CREATE INDEX IF NOT EXISTS idx_orders_company_id ON orders(company_id);
    `

    _, err := DB.Exec(query)
    if err != nil {
        log.Fatalf("migration failed: %v", err)
    }
}