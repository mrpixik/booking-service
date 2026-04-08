-- Пользователи
INSERT INTO users (id, email, role, password_hash) VALUES
                                                       ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@example.com', 'admin', '$2a$10$dummyhashadmin'),
                                                       ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'alice@example.com', 'user', '$2a$10$dummyhashalice'),
                                                       ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'bob@example.com', 'user', '$2a$10$dummyhashbob')
ON CONFLICT DO NOTHING;

-- Комнаты
INSERT INTO rooms (id, name, description, capacity) VALUES
                                                        ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'Переговорная А', 'Маленькая переговорная на 2 этаже', 4),
                                                        ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 'Конференц-зал Б', 'Большой зал с проектором', 20),
                                                        ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', 'Комната С', 'Тихая комната для звонков', 2)
ON CONFLICT DO NOTHING;

-- Расписания (пн-пт = {1,2,3,4,5})
INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time) VALUES
                                                                            ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', '{1,2,3,4,5}', '09:00', '18:00'),
                                                                            ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', '{1,2,3,4,5}', '08:00', '20:00'),
                                                                            ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', '{1,2,3,4,5,6}', '10:00', '17:00')
ON CONFLICT DO NOTHING;

-- Слоты на 27 апреля 2026 (понедельник)
INSERT INTO slots (id, room_id, start_at, end_at) VALUES
                                                      ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', '2026-04-27 09:00:00+03', '2026-04-27 10:00:00+03'),
                                                      ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', '2026-04-27 10:00:00+03', '2026-04-27 11:00:00+03'),
                                                      ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', '2026-04-27 11:00:00+03', '2026-04-27 12:00:00+03'),
                                                      ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a04', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', '2026-04-27 08:00:00+03', '2026-04-27 09:00:00+03'),
                                                      ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a05', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', '2026-04-27 09:00:00+03', '2026-04-27 10:00:00+03'),
                                                      ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a06', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', '2026-04-27 10:00:00+03', '2026-04-27 11:00:00+03')
ON CONFLICT DO NOTHING;

-- Бронирования
INSERT INTO bookings (id, slot_id, user_id, status, conference_link) VALUES
                                                                         ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12', 'active', 'https://meet.example.com/abc123'),
                                                                         ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a04', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'active', 'https://meet.example.com/def456'),
                                                                         ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a03', 'd0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13', 'cancelled', NULL)
ON CONFLICT DO NOTHING;