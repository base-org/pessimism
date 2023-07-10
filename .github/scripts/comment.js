/*
  This script will create a comment on the pull request with the current test coverage.
  If the comment already exists, it will update the existing comment.

  This script is called from the .github/workflows/test.yml workflow file. It is very difficult to degug
  given the way GitHub Actions works.

*/
module.exports = async ({github, context, core}) => {
  const fs = require('fs');
  const coverage = fs.readFileSync('out.txt', 'utf8');
  const reportBody = '### Current Test Coverage\n' + '```' + coverage + '```';

  if (process.env.DATA == null || process.env.DATA == ''){
    // Create a comment on pull request
    github.rest.issues.createComment({
        issue_number: context.issue.number,
        owner: context.repo.owner,
        repo: context.repo.repo,
        body: reportBody
      })

  } else {
    // Update existing comment on pull request
    github.rest.issues.updateComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        comment_id: process.env.DATA,
        body: '### Current Test Coverage\n' + '```' + coverage + '```'
      })
  }
}
