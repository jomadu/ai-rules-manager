# Task List Management

Guidelines for managing task lists in markdown files to track progress on completing a PRD

## Task Implementation
- **One sub-task at a time:** Do **NOT** start the next sub‑task until you ask the user for permission and they say "yes" or "y"
- **Completion protocol:**
  1. When you finish a **task**, follow this sequence:
    - **First**: Mark it as completed by changing `[ ]` to `[x]`
    - **Run tests**: Run relevant tests for the subtask (`pytest path/to/test`, `npm test -- path/to/test`, etc.)
    - **Only if tests pass**: Stage changes (`git add .`)
    - **Clean up**: Remove any temporary files and temporary code before committing
    - **Commit**: Use a descriptive commit message that:
      - Uses conventional commit format (`feat:`, `fix:`, `refactor:`, etc.)
      - Summarizes what was accomplished in the subtask
      - Lists key changes and additions
      - References the subtask number and PRD context
      - **Formats the message as a single-line command using `-m` flags**, e.g.:

        ```
        git commit -m "feat: implement user validation" -m "- Add email format validation" -m "- Include unit tests" -m "Subtask 1.2 from PRD"
        ```
  2. Once all the subtasks are marked completed, mark the **parent task** as completed.
- Stop after each sub‑task and wait for the user's go‑ahead.

## Task List Maintenance

1. **Update the task list as you work:**
   - Mark tasks and subtasks as completed (`[x]`) per the protocol above.
   - Add new tasks as they emerge.

2. **Maintain the "Relevant Files" section:**
   - List every file created or modified.
   - Give each file a one‑line description of its purpose.

## AI Instructions

When working with task lists, the AI must:

1. **Consider additional context**: Before starting any sub-task, review all files and subdirectories in the same directory as the PRD (e.g., technical specifications, example workflows, design documents) to ensure implementation aligns with all available context.
2. Regularly update the task list file after finishing any significant work.
3. Follow the completion protocol:
   - Mark each finished **sub‑task** `[x]`.
   - Mark the **parent task** `[x]` once **all** its subtasks are `[x]`.
4. Add newly discovered tasks.
5. Keep "Relevant Files" accurate and up to date.
6. Before starting work, check which sub‑task is next.
7. After implementing a sub‑task, update the file and then pause for user approval.
