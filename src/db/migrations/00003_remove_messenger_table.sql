-- +goose Up
-- El módulo `messenger` (cmd/messenger, messenger/*) era una refactorización
-- que nunca se terminó de implementar ni se llegó a desplegar. Se eliminó el
-- código; esta tabla era la única cosa que usaba.
DROP TABLE IF EXISTS Messages;

-- +goose Down
CREATE TABLE Messages (
    id           BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    phone        CHAR(16)    NOT NULL,
    type         VARCHAR(32) NOT NULL DEFAULT 'text',
    content      TEXT        NOT NULL,
    outgoing     BOOLEAN     NOT NULL DEFAULT FALSE,
    scheduled_at DATETIME    NOT NULL,
    sended_at    DATETIME    DEFAULT NULL,
    on_response  CHAR(37)    DEFAULT NULL,

    INDEX idx_phone   (phone),
    INDEX idx_pending (outgoing, scheduled_at, sended_at)
);
