export namespace main {
	
	export class DefaultPaths {
	    minecraftLauncherPath: string;
	    minecraftSavePath: string;
	
	    static createFrom(source: any = {}) {
	        return new DefaultPaths(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.minecraftLauncherPath = source["minecraftLauncherPath"];
	        this.minecraftSavePath = source["minecraftSavePath"];
	    }
	}
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

