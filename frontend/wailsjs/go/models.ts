export namespace app {
	
	export class Info {
	    name: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	    }
	}

}

export namespace calendar {
	
	export class DueReminder {
	    eventId: string;
	    eventTitle: string;
	    reminderId: string;
	    occurrenceStart: time.Time;
	    fireTime: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new DueReminder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eventId = source["eventId"];
	        this.eventTitle = source["eventTitle"];
	        this.reminderId = source["reminderId"];
	        this.occurrenceStart = this.convertValues(source["occurrenceStart"], time.Time);
	        this.fireTime = this.convertValues(source["fireTime"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Reminder {
	    id: string;
	    leadTime: number;
	
	    static createFrom(source: any = {}) {
	        return new Reminder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.leadTime = source["leadTime"];
	    }
	}
	export class Exception {
	    originalDate: time.Time;
	    type: string;
	    newStart?: time.Time;
	    newEnd?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Exception(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.originalDate = this.convertValues(source["originalDate"], time.Time);
	        this.type = source["type"];
	        this.newStart = this.convertValues(source["newStart"], time.Time);
	        this.newEnd = this.convertValues(source["newEnd"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RecurrenceRule {
	    frequency: string;
	    interval: number;
	    count?: number;
	    until?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new RecurrenceRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.frequency = source["frequency"];
	        this.interval = source["interval"];
	        this.count = source["count"];
	        this.until = this.convertValues(source["until"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Event {
	    id: string;
	    title: string;
	    description?: string;
	    location?: string;
	    start: time.Time;
	    end: time.Time;
	    allDay: boolean;
	    important: boolean;
	    recurrence?: RecurrenceRule;
	    exceptions?: Exception[];
	    reminders?: Reminder[];
	    createdAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.location = source["location"];
	        this.start = this.convertValues(source["start"], time.Time);
	        this.end = this.convertValues(source["end"], time.Time);
	        this.allDay = source["allDay"];
	        this.important = source["important"];
	        this.recurrence = this.convertValues(source["recurrence"], RecurrenceRule);
	        this.exceptions = this.convertValues(source["exceptions"], Exception);
	        this.reminders = this.convertValues(source["reminders"], Reminder);
	        this.createdAt = this.convertValues(source["createdAt"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class OccurrenceView {
	    start: time.Time;
	    end: time.Time;
	    originalStart: time.Time;
	    eventId: string;
	    title: string;
	    description?: string;
	    location?: string;
	    allDay: boolean;
	    important: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OccurrenceView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = this.convertValues(source["start"], time.Time);
	        this.end = this.convertValues(source["end"], time.Time);
	        this.originalStart = this.convertValues(source["originalStart"], time.Time);
	        this.eventId = source["eventId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.location = source["location"];
	        this.allDay = source["allDay"];
	        this.important = source["important"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace tasks {
	
	export class Task {
	    id: string;
	    title: string;
	    done: boolean;
	    priority: string;
	    createdAt: time.Time;
	    completedAt?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.done = source["done"];
	        this.priority = source["priority"];
	        this.createdAt = this.convertValues(source["createdAt"], time.Time);
	        this.completedAt = this.convertValues(source["completedAt"], time.Time);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace time {
	
	export class Time {
	
	
	    static createFrom(source: any = {}) {
	        return new Time(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

