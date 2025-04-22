export namespace main {
	
	export class UserData {
	    minecraftLauncher: string;
	    minecraftDirectory: string;
	    worldName: string;
	
	    static createFrom(source: any = {}) {
	        return new UserData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minecraftLauncher = source["minecraftLauncher"];
	        this.minecraftDirectory = source["minecraftDirectory"];
	        this.worldName = source["worldName"];
	    }
	}

}

