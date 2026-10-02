-- Level 5 (Growth): lessons, quizzes, questions, options.
INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES
    (
        '00000000-0000-0000-0000-000000004101',
        '00000000-0000-0000-0000-000000000005',
        'What Resilience Really Means',
        'Resilience is often misunderstood as "not being affected" by hard things, or "bouncing back" as though nothing happened. A more accurate picture is that resilience is the capacity to be affected by a difficult experience and still find a way forward -- it includes the struggle, not an absence of it.\n\nResilience is also not a fixed personal trait some people have and others lack. It is built through a combination of factors: supportive relationships, practical coping skills, a sense of meaning, and often, practice recovering from smaller setbacks before facing larger ones.\n\nFraming resilience this way matters because it means it can be developed. Struggling with something difficult is not evidence that you lack resilience -- it may be the process resilience actually looks like while it''s being built.',
        1,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000004102',
        '00000000-0000-0000-0000-000000000005',
        'Clarifying Your Values',
        'Values are the qualities and priorities that matter most to how you want to live -- things like honesty, connection, growth, or independence. They are different from goals: a goal is something you complete, while a value is a direction you keep moving toward, never fully "finished."\n\nAfter trauma, it''s common to feel disconnected from what you value, especially if survival took priority for a long time. Revisiting the question "what actually matters to me" -- not what should matter, or what mattered to someone else -- can be a grounding and clarifying exercise.\n\nOne simple approach is noticing moments that felt meaningful or energizing in the past, even small ones, and asking what value they reflected. Over time, this builds a clearer picture to guide decisions, especially when a choice feels confusing or directionless.',
        2,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000004103',
        '00000000-0000-0000-0000-000000000005',
        'Trauma-Informed Goal Setting',
        'Goal setting after trauma works best when it accounts for how recovery actually happens: non-linearly, with good days and setbacks, rather than steady upward progress. Goals that assume constant improvement often set people up to feel like they''ve failed when a harder week happens -- which is a normal part of the process, not a sign of regression.\n\nSmaller, process-focused goals ("I will try one grounding technique this week") tend to be more sustainable than large outcome-focused ones ("I will no longer feel anxious"), because they are within your direct control and give you something to actually act on.\n\nIt also helps to build in flexibility: a goal can be adjusted without being abandoned. Revisiting and resetting a goal is part of working with it, not a sign the goal -- or you -- failed.',
        3,
        UTC_TIMESTAMP()
    );

INSERT INTO quizzes (id, lesson_id, created_at) VALUES
    ('00000000-0000-0000-0000-000000004201', '00000000-0000-0000-0000-000000004101', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000004202', '00000000-0000-0000-0000-000000004102', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000004203', '00000000-0000-0000-0000-000000004103', UTC_TIMESTAMP());

INSERT INTO quiz_questions (id, quiz_id, question, order_index) VALUES
    ('00000000-0000-0000-0000-000000004301', '00000000-0000-0000-0000-000000004201', 'According to this lesson, what does resilience include?', 1),
    ('00000000-0000-0000-0000-000000004302', '00000000-0000-0000-0000-000000004202', 'How does this lesson describe the difference between a value and a goal?', 1),
    ('00000000-0000-0000-0000-000000004303', '00000000-0000-0000-0000-000000004203', 'Why does this lesson recommend process-focused goals over outcome-focused ones?', 1);

INSERT INTO quiz_options (id, question_id, text, is_correct, order_index) VALUES
    ('00000000-0000-0000-0000-000000004401', '00000000-0000-0000-0000-000000004301', 'Being completely unaffected by hard things', FALSE, 1),
    ('00000000-0000-0000-0000-000000004402', '00000000-0000-0000-0000-000000004301', 'The struggle itself, not an absence of it', TRUE, 2),
    ('00000000-0000-0000-0000-000000004403', '00000000-0000-0000-0000-000000004301', 'A fixed trait you either have or don''t', FALSE, 3),
    ('00000000-0000-0000-0000-000000004404', '00000000-0000-0000-0000-000000004301', 'Avoiding difficult situations entirely', FALSE, 4),

    ('00000000-0000-0000-0000-000000004411', '00000000-0000-0000-0000-000000004302', 'A goal is a direction you move toward forever; a value is something you complete', FALSE, 1),
    ('00000000-0000-0000-0000-000000004412', '00000000-0000-0000-0000-000000004302', 'A value is a direction you keep moving toward; a goal is something you complete', TRUE, 2),
    ('00000000-0000-0000-0000-000000004413', '00000000-0000-0000-0000-000000004302', 'They mean exactly the same thing', FALSE, 3),
    ('00000000-0000-0000-0000-000000004414', '00000000-0000-0000-0000-000000004302', 'Values only apply to career decisions', FALSE, 4),

    ('00000000-0000-0000-0000-000000004421', '00000000-0000-0000-0000-000000004303', 'They are within your direct control and give you something to act on', TRUE, 1),
    ('00000000-0000-0000-0000-000000004422', '00000000-0000-0000-0000-000000004303', 'They guarantee a specific outcome', FALSE, 2),
    ('00000000-0000-0000-0000-000000004423', '00000000-0000-0000-0000-000000004303', 'They are easier to brag about to others', FALSE, 3),
    ('00000000-0000-0000-0000-000000004424', '00000000-0000-0000-0000-000000004303', 'They require no flexibility once set', FALSE, 4);
