-- Level 2 (Self-Awareness): lessons, quizzes, questions, options. The
-- level row and its badge already exist (migration 0014) — only
-- content was missing.
INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES
    (
        '00000000-0000-0000-0000-000000001101',
        '00000000-0000-0000-0000-000000000002',
        'The Four Trauma Responses',
        'When your nervous system senses a threat, it doesn''t wait for permission to react. Most people have heard of "fight or flight," but there are actually four common patterns: fight (confronting the threat), flight (escaping it), freeze (shutting down, unable to move or speak), and fawn (appeasing or people-pleasing to stay safe).\n\nNone of these is "wrong" or "weak." Each one is a survival strategy your body learned works in a given situation, often shaped by experiences earlier in your life. Someone raised in an unpredictable household might default to fawning -- scanning a room for tension and smoothing it over -- long after that environment is gone, because the pattern kept them safe once.\n\nRecognizing which response you tend toward isn''t about judging yourself. It''s the first step in noticing when your body is reacting to a present-day situation versus replaying an old survival pattern that no longer fits.',
        1,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000001102',
        '00000000-0000-0000-0000-000000000002',
        'Understanding Your Triggers',
        'A trigger is anything -- a sound, a smell, a tone of voice, a specific phrase -- that causes your nervous system to react as though a past threat is happening again, even when you are currently safe. Triggers don''t need to make logical sense to be real; they are your body''s associative memory, not a reasoned argument.\n\nBeing triggered often feels disproportionate to what just happened, which can be confusing or embarrassing. It helps to remember that the intensity of the reaction belongs to the original experience the trigger is connected to, not necessarily to the present moment.\n\nA simple starting practice is noticing, after a strong reaction, three things: what happened right before it, what you felt in your body, and what thought came with it. Over time this builds a map of your own triggers -- not to avoid life around them, but to recognize "this is a trigger" in the moment instead of "something is wrong with me."',
        2,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000001103',
        '00000000-0000-0000-0000-000000000002',
        'Building a Personal Awareness Practice',
        'Self-awareness around trauma responses is a skill, not a one-time insight -- it develops through small, repeated check-ins rather than a single breakthrough. A useful habit is pausing a few times a day to ask: what am I feeling in my body right now, and what state am I in (calm, activated, shut down)?\n\nJournaling (see this app''s Journal tool) and mood tracking are both practical ways to build this awareness over time, because they turn vague "I felt off today" into specific, reviewable patterns -- certain times of day, certain people, certain situations.\n\nThis practice is not about achieving constant calm. It''s about shortening the gap between a reaction happening and recognizing it, so you have more choice in how you respond next.',
        3,
        UTC_TIMESTAMP()
    );

INSERT INTO quizzes (id, lesson_id, created_at) VALUES
    ('00000000-0000-0000-0000-000000001201', '00000000-0000-0000-0000-000000001101', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000001202', '00000000-0000-0000-0000-000000001102', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000001203', '00000000-0000-0000-0000-000000001103', UTC_TIMESTAMP());

INSERT INTO quiz_questions (id, quiz_id, question, order_index) VALUES
    ('00000000-0000-0000-0000-000000001301', '00000000-0000-0000-0000-000000001201', 'Besides fight and flight, what are the other two common trauma responses?', 1),
    ('00000000-0000-0000-0000-000000001302', '00000000-0000-0000-0000-000000001202', 'What determines the intensity of a triggered reaction?', 1),
    ('00000000-0000-0000-0000-000000001303', '00000000-0000-0000-0000-000000001203', 'What is the goal of a personal awareness practice, according to this lesson?', 1);

INSERT INTO quiz_options (id, question_id, text, is_correct, order_index) VALUES
    ('00000000-0000-0000-0000-000000001401', '00000000-0000-0000-0000-000000001301', 'Freeze and fawn', TRUE, 1),
    ('00000000-0000-0000-0000-000000001402', '00000000-0000-0000-0000-000000001301', 'Panic and rest', FALSE, 2),
    ('00000000-0000-0000-0000-000000001403', '00000000-0000-0000-0000-000000001301', 'Avoidance and anger', FALSE, 3),
    ('00000000-0000-0000-0000-000000001404', '00000000-0000-0000-0000-000000001301', 'Sleep and appetite change', FALSE, 4),

    ('00000000-0000-0000-0000-000000001411', '00000000-0000-0000-0000-000000001302', 'How socially acceptable the reaction looks to others', FALSE, 1),
    ('00000000-0000-0000-0000-000000001412', '00000000-0000-0000-0000-000000001302', 'The original experience the trigger is connected to', TRUE, 2),
    ('00000000-0000-0000-0000-000000001413', '00000000-0000-0000-0000-000000001302', 'How much sleep you got the night before', FALSE, 3),
    ('00000000-0000-0000-0000-000000001414', '00000000-0000-0000-0000-000000001302', 'Nothing -- the reaction is always random', FALSE, 4),

    ('00000000-0000-0000-0000-000000001421', '00000000-0000-0000-0000-000000001303', 'Achieving constant calm at all times', FALSE, 1),
    ('00000000-0000-0000-0000-000000001422', '00000000-0000-0000-0000-000000001303', 'Shortening the gap between a reaction and recognizing it', TRUE, 2),
    ('00000000-0000-0000-0000-000000001423', '00000000-0000-0000-0000-000000001303', 'Avoiding all situations that could cause a reaction', FALSE, 3),
    ('00000000-0000-0000-0000-000000001424', '00000000-0000-0000-0000-000000001303', 'Proving to others that you are not overreacting', FALSE, 4);
