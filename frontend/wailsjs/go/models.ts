export namespace main {
	
	export class UserData {
	    minecraftDirectory: string;
	    worldName: string;
	
	    static createFrom(source: any = {}) {
	        return new UserData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minecraftDirectory = source["minecraftDirectory"];
	        this.worldName = source["worldName"];
	    }
	}

}

