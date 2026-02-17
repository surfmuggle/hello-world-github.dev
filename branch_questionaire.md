# Branch: questionaire

Prompt given:

1. Create a branch questionaire
2. add a navigation on the left side of the html page with a max width of 40rem - put the navigation into it own template nav.html file 
3. Show how render form elements (input, select, text, textearea, radio) based on a json file that contains title, description, formtype (radio, input text, input radio, input checkbox, texarea). 
4. store the form values als json into a new table questions inside sqlite
5. write this prompt and your actions into branch_questionaire.md (See <attachments> above for file contents. You may not need to search or read the file again.)

What I changed and added:

- Created and switched to branch `questionaire`.
- Updated `main.go`:
  - Added a new `questions` table creation in SQLite.
  - Added `questionnaireHandler` to read `data/questions.json`, build HTML form elements dynamically, and render `templates/questionnaire.html`.
  - Added `submitQuestionnaireHandler` to accept POST submissions, encode form values as JSON, and insert into the `questions` table.
  - Parsed `nav.html` into templates so it can be reused by pages.
- Templates:
  - Added `templates/nav.html` (left navigation, max-width 40rem via CSS).
  - Added `templates/questionnaire.html` (page that receives rendered form parts and posts to `/submit-questionnaire`).
  - Updated `templates/home.html` to include the left navigation in a two-column layout.
- Added `data/questions.json` as a sample JSON form definition demonstrating `input text`, `input radio`, `input checkbox`, and `textarea`.

How the dynamic form rendering works:

- The server reads `data/questions.json` which is an array of objects with `title`, `description`, `formtype`, `name`, and optional `options`.
- For each field the server creates a small HTML snippet (as `template.HTML`) depending on `formtype`:
  - `input text` -> `<input type="text">`
  - `textarea` -> `<textarea>`
  - `input radio` -> series of `<input type="radio">` for each `options` value
  - `input checkbox` -> series of `<input type="checkbox">` for each `options` value
  - `select` -> `<select>` with `<option>` entries
- Those snippets are passed to `templates/questionnaire.html` and rendered inside the form.

How submissions are stored:

- On POST to `/submit-questionnaire` the server collects `r.PostForm`, converts single vs multiple values appropriately to a map, marshals it as JSON, and inserts into the `questions` table in `entries.db`.

Files added/modified:

- Modified: `main.go`
- Modified: `templates/home.html`
- Added: `templates/nav.html`
- Added: `templates/questionnaire.html`
- Added: `data/questions.json`
- Added: `branch_questionaire.md` (this file)

Next steps you might want:

- Show a listing view for stored questionnaire responses.
- Add validation and CSRF protection.
- Improve styling and mobile responsiveness.

