-- Level 6 (Advanced): lessons, quizzes, questions, options.
INSERT INTO lessons (id, level_id, title, content, order_index, created_at) VALUES
    (
        '00000000-0000-0000-0000-000000005101',
        '00000000-0000-0000-0000-000000000006',
        'Principles of Trauma-Informed Leadership',
        'Trauma-informed leadership means recognizing that the people you lead -- whether as a manager, a peer supporter, a teacher, or a community organizer -- may be carrying experiences you cannot see, and adjusting how you lead accordingly. It is not about diagnosing anyone or becoming their therapist; it is about reducing unnecessary harm and creating conditions where people can function and contribute.\n\nFour widely cited principles are: safety (physical and emotional), trustworthiness and transparency (consistent, predictable communication), choice and collaboration (involving people in decisions that affect them rather than imposing on them), and empowerment (building on strengths rather than focusing only on deficits).\n\nApplying these doesn''t require special credentials. Small, consistent practices -- clear expectations, following through on what you say, asking rather than assuming -- go a long way toward making a space feel safer for everyone in it, including people whose history you''ll never know.',
        1,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000005102',
        '00000000-0000-0000-0000-000000000006',
        'Supporting Others Without Overstepping',
        'Supporting someone who has experienced trauma is different from treating them. Peer and informal support is valuable precisely because it offers connection and understanding -- not a clinical intervention. Knowing that distinction, and where your role ends, protects both you and the person you''re supporting.\n\nUseful support often looks like listening without trying to fix, believing what someone shares without demanding details or proof, and not pressuring anyone to "move on" on a timeline that suits you rather than them. It also means recognizing when a situation is beyond what peer support can safely address, and knowing how to point someone toward professional help without abandoning them in the process.\n\nOverstepping -- giving advice as if it were clinical guidance, pushing someone to share more than they''re ready to, or taking on responsibility for someone else''s healing -- can cause real harm, even with good intentions. Good support holds space; it doesn''t take over.',
        2,
        UTC_TIMESTAMP()
    ),
    (
        '00000000-0000-0000-0000-000000005103',
        '00000000-0000-0000-0000-000000000006',
        'Sustaining Yourself While Supporting Others',
        'Supporting people through difficult experiences, whether formally or informally, carries a real cost if you don''t also tend to yourself. Vicarious or secondary trauma -- absorbing distress from repeatedly hearing about others'' difficult experiences -- is a recognized risk for caregivers, peer supporters, and leaders alike, not a sign you are unsuited for the role.\n\nSustainable support requires the same regulation and awareness practices covered earlier in this course, applied to yourself: noticing your own capacity, setting boundaries around what you take on, and having your own sources of support rather than only being a source of support for others.\n\nThere is no version of this work that is sustainable without rest and limits. Protecting your own capacity is not selfish or a failure of commitment -- it is what allows you to keep showing up for others without burning out or causing harm through depletion.',
        3,
        UTC_TIMESTAMP()
    );

INSERT INTO quizzes (id, lesson_id, created_at) VALUES
    ('00000000-0000-0000-0000-000000005201', '00000000-0000-0000-0000-000000005101', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000005202', '00000000-0000-0000-0000-000000005102', UTC_TIMESTAMP()),
    ('00000000-0000-0000-0000-000000005203', '00000000-0000-0000-0000-000000005103', UTC_TIMESTAMP());

INSERT INTO quiz_questions (id, quiz_id, question, order_index) VALUES
    ('00000000-0000-0000-0000-000000005301', '00000000-0000-0000-0000-000000005201', 'Which of these is one of the four trauma-informed leadership principles described in this lesson?', 1),
    ('00000000-0000-0000-0000-000000005302', '00000000-0000-0000-0000-000000005202', 'What is an example of "overstepping" when supporting someone, per this lesson?', 1),
    ('00000000-0000-0000-0000-000000005303', '00000000-0000-0000-0000-000000005203', 'What does this lesson call the risk of absorbing distress from repeatedly supporting others?', 1);

INSERT INTO quiz_options (id, question_id, text, is_correct, order_index) VALUES
    ('00000000-0000-0000-0000-000000005401', '00000000-0000-0000-0000-000000005301', 'Choice and collaboration', TRUE, 1),
    ('00000000-0000-0000-0000-000000005402', '00000000-0000-0000-0000-000000005301', 'Strict obedience', FALSE, 2),
    ('00000000-0000-0000-0000-000000005403', '00000000-0000-0000-0000-000000005301', 'Maximum efficiency', FALSE, 3),
    ('00000000-0000-0000-0000-000000005404', '00000000-0000-0000-0000-000000005301', 'Competitive ranking', FALSE, 4),

    ('00000000-0000-0000-0000-000000005411', '00000000-0000-0000-0000-000000005302', 'Listening without trying to fix the situation', FALSE, 1),
    ('00000000-0000-0000-0000-000000005412', '00000000-0000-0000-0000-000000005302', 'Giving advice as if it were clinical guidance', TRUE, 2),
    ('00000000-0000-0000-0000-000000005413', '00000000-0000-0000-0000-000000005302', 'Pointing someone toward professional help when needed', FALSE, 3),
    ('00000000-0000-0000-0000-000000005414', '00000000-0000-0000-0000-000000005302', 'Believing what someone shares without demanding proof', FALSE, 4),

    ('00000000-0000-0000-0000-000000005421', '00000000-0000-0000-0000-000000005303', 'Vicarious or secondary trauma', TRUE, 1),
    ('00000000-0000-0000-0000-000000005422', '00000000-0000-0000-0000-000000005303', 'Professional burnout only', FALSE, 2),
    ('00000000-0000-0000-0000-000000005423', '00000000-0000-0000-0000-000000005303', 'A sign someone is unsuited for the role', FALSE, 3),
    ('00000000-0000-0000-0000-000000005424', '00000000-0000-0000-0000-000000005303', 'An unavoidable, untreatable condition', FALSE, 4);
