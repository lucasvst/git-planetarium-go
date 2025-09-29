export namespace git_manager {
	
	export class RepositoryInfo {
	    name: string;
	    last_commit_date: string;
	    branch_count: number;
	
	    static createFrom(source: any = {}) {
	        return new RepositoryInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.last_commit_date = source["last_commit_date"];
	        this.branch_count = source["branch_count"];
	    }
	}

}

