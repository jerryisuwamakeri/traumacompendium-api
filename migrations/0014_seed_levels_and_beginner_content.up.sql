-- All six levels from the brief's learning framework, so the full
-- roadmap is browsable even though only Level 1 has lesson content yet.
INSERT INTO levels (id, number, name, description, created_at) VALUES
    ('00000000-0000-0000-0000-000000000001', 1, 'Beginner', 'Trauma basics, myths, brain science', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000002', 2, 'Self-Awareness', 'Fight/flight/freeze/fawn, triggers', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000003', 3, 'Regulation', 'Grounding, sleep, stress management', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000004', 4, 'Relationships', 'Boundaries, trust, attachment styles', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000005', 5, 'Growth', 'Resilience, values, goal setting', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000006', 6, 'Advanced', 'Trauma-informed leadership & peer support', UTC_TIMESTAMP());

-- Level 1 lessons.
INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES
    (
        '00000000-0000-0000-0000-000000000101',
        '00000000-0000-0000-0000-000000000001',
        'What Is Trauma?',
        'Trauma is your nervous system''s response to an event (or series of events) that overwhelmed your ability to cope at the time. It is not about how severe an event looks from the outside -- it is about how your body and mind experienced it.\n\nTrauma can come from a single event (an accident, an assault, a loss) or from repeated experiences over time (ongoing stress, neglect, an unsafe environment). Two people can go through the same event and have very different responses, because trauma is shaped by the event itself, your history, your support system, and your nervous system''s wiring.\n\nThis platform does not diagnose or treat trauma. Think of these lessons as a starting vocabulary for understanding your own experience -- a foundation you can build on with a qualified professional if and when you choose to.',
        1,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000000102',
        '00000000-0000-0000-0000-000000000001',
        'Common Myths About Trauma',
        'Myth: "Only catastrophic events (war, disasters) cause trauma."\nReality: Trauma can come from many kinds of experiences, including ones that seem "ordinary" from the outside, like chronic stress, bullying, or an unstable home.\n\nMyth: "If you don''t remember it clearly, it wasn''t that bad."\nReality: Memory under extreme stress often works differently -- fragmented or fuzzy memories are common and do not mean an experience was minor.\n\nMyth: "Strong people don''t get affected by trauma."\nReality: Trauma responses are a normal nervous system reaction, not a character flaw or a sign of weakness.\n\nMyth: "Time alone heals trauma."\nReality: Time can help, but understanding, support, and sometimes professional care are usually what actually move someone from surviving to healing.',
        2,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000000103',
        '00000000-0000-0000-0000-000000000001',
        'How Trauma Affects the Brain',
        'When your brain senses danger, it prioritizes survival over reflection. A region called the amygdala acts like a smoke alarm, triggering a stress response before your slower, more reflective thinking (in the prefrontal cortex) has a chance to fully weigh in.\n\nThis is why, in a frightening moment, people often react instantly rather than "thinking it through" -- that is the nervous system doing exactly what it evolved to do.\n\nWith repeated or prolonged stress, this alarm system can become more sensitive, staying "on alert" even in safe situations. That is not a flaw in you; it is a learned pattern that can also be unlearned with time, safety, and support. Later levels in this course go deeper into recognizing and working with these responses.',
        3,
        UTC_TIMESTAMP()
    );

-- Level 1 quizzes (one per lesson) with questions and options.
INSERT INTO quizzes (id, lesson_id, created_at) VALUES
    ('00000000-0000-0000-0000-000000000201', '00000000-0000-0000-0000-000000000101', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000202', '00000000-0000-0000-0000-000000000102', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000000203', '00000000-0000-0000-0000-000000000103', UTC_TIMESTAMP());

INSERT INTO quiz_questions (id, quiz_id, question, order_index) VALUES
    ('00000000-0000-0000-0000-000000000301', '00000000-0000-0000-0000-000000000201', 'What determines whether an event is traumatic for someone?', 1),
    ('00000000-0000-0000-0000-000000000302', '00000000-0000-0000-0000-000000000202', 'Which statement about trauma is accurate?', 1),
    ('00000000-0000-0000-0000-000000000303', '00000000-0000-0000-0000-000000000203', 'What is the amygdala often compared to?', 1);

INSERT INTO quiz_options (id, question_id, text, is_correct, order_index) VALUES
    ('00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000301', 'Only how severe the event looks to other people', FALSE, 1),
    ('00000000-0000-0000-0000-000000000402', '00000000-0000-0000-0000-000000000301', 'How your body and mind experienced it at the time', TRUE, 2),
    ('00000000-0000-0000-0000-000000000403', '00000000-0000-0000-0000-000000000301', 'Whether it happened more than once', FALSE, 3),
    ('00000000-0000-0000-0000-000000000404', '00000000-0000-0000-0000-000000000301', 'How long ago it happened', FALSE, 4),

    ('00000000-0000-0000-0000-000000000411', '00000000-0000-0000-0000-000000000302', 'Trauma responses are a sign of weakness', FALSE, 1),
    ('00000000-0000-0000-0000-000000000412', '00000000-0000-0000-0000-000000000302', 'Trauma can only come from catastrophic events', FALSE, 2),
    ('00000000-0000-0000-0000-000000000413', '00000000-0000-0000-0000-000000000302', 'Fragmented memories are a common, normal response to extreme stress', TRUE, 3),
    ('00000000-0000-0000-0000-000000000414', '00000000-0000-0000-0000-000000000302', 'Time alone is always enough to heal trauma', FALSE, 4),

    ('00000000-0000-0000-0000-000000000421', '00000000-0000-0000-0000-000000000303', 'A library', FALSE, 1),
    ('00000000-0000-0000-0000-000000000422', '00000000-0000-0000-0000-000000000303', 'A smoke alarm', TRUE, 2),
    ('00000000-0000-0000-0000-000000000423', '00000000-0000-0000-0000-000000000303', 'A thermostat', FALSE, 3),
    ('00000000-0000-0000-0000-000000000424', '00000000-0000-0000-0000-000000000303', 'A filing cabinet', FALSE, 4);

-- One badge per level; only the Beginner badge is reachable until later
-- levels get lesson content.
INSERT INTO badges (id, level_id, name, description, icon_key) VALUES
    ('00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000001', 'Beginner', 'Completed the Beginner level: trauma basics, myths, and brain science.', 'badge-beginner'),
    ('00000000-0000-0000-0000-000000000502', '00000000-0000-0000-0000-000000000002', 'Self-Aware', 'Completed the Self-Awareness level.', 'badge-self-awareness'),
    ('00000000-0000-0000-0000-000000000503', '00000000-0000-0000-0000-000000000003', 'Regulated', 'Completed the Regulation level.', 'badge-regulation'),
    ('00000000-0000-0000-0000-000000000504', '00000000-0000-0000-0000-000000000004', 'Connected', 'Completed the Relationships level.', 'badge-relationships'),
    ('00000000-0000-0000-0000-000000000505', '00000000-0000-0000-0000-000000000005', 'Growing', 'Completed the Growth level.', 'badge-growth'),
    ('00000000-0000-0000-0000-000000000506', '00000000-0000-0000-0000-000000000006', 'Trauma-Informed Leader', 'Completed the Advanced level.', 'badge-advanced');
