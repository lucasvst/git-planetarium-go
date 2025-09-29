export interface Repository {
  name: string;
  last_commit_date: string;
  branch_count: number;
}

export {
  Setup,
  ListRepositories,
  GitClone,
} from "./../../../wailsjs/go/git_manager/GitManager";