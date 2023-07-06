


module.exports = async ({github, context, core}) => {
  const fs = require('fs');
  const coverage = fs.readFileSync('out.txt', 'utf8');
  const reportBody = '### Current Test Coverage\n' + '```' + coverage + '```';

  if (process.env.DATA != null || process.env.DATA == ''){
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
