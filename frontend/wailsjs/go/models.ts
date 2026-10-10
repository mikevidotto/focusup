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
	
	export class Completion {
	    occurrenceDate: time.Time;
	    completedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Completion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.occurrenceDate = this.convertValues(source["occurrenceDate"], time.Time);
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
	    weekdays?: number[];
	    count?: number;
	    until?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new RecurrenceRule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.frequency = source["frequency"];
	        this.interval = source["interval"];
	        this.weekdays = source["weekdays"];
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
	    completions?: Completion[];
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
	        this.completions = this.convertValues(source["completions"], Completion);
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
	export class EventInput {
	    title: string;
	    description: string;
	    location: string;
	    start: time.Time;
	    end: time.Time;
	    allDay: boolean;
	    important: boolean;
	    recurrence?: RecurrenceRule;
	    reminderLeadSeconds: number[];
	
	    static createFrom(source: any = {}) {
	        return new EventInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.description = source["description"];
	        this.location = source["location"];
	        this.start = this.convertValues(source["start"], time.Time);
	        this.end = this.convertValues(source["end"], time.Time);
	        this.allDay = source["allDay"];
	        this.important = source["important"];
	        this.recurrence = this.convertValues(source["recurrence"], RecurrenceRule);
	        this.reminderLeadSeconds = source["reminderLeadSeconds"];
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
	    done: boolean;
	    eventId: string;
	    title: string;
	    description?: string;
	    location?: string;
	    allDay: boolean;
	    important: boolean;
	    recurring: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OccurrenceView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = this.convertValues(source["start"], time.Time);
	        this.end = this.convertValues(source["end"], time.Time);
	        this.originalStart = this.convertValues(source["originalStart"], time.Time);
	        this.done = source["done"];
	        this.eventId = source["eventId"];
	        this.title = source["title"];
	        this.description = source["description"];
	        this.location = source["location"];
	        this.allDay = source["allDay"];
	        this.important = source["important"];
	        this.recurring = source["recurring"];
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

export namespace habits {
	
	export class Habit {
	    id: string;
	    name: string;
	    createdAt: time.Time;
	    completions: string[];
	
	    static createFrom(source: any = {}) {
	        return new Habit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = this.convertValues(source["createdAt"], time.Time);
	        this.completions = source["completions"];
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

export namespace jobs {
	
	export class Application {
	    row: number;
	    date: string;
	    company: string;
	    sector: string;
	    role: string;
	    roleType: string;
	    channel: string;
	    status: string;
	    contactPerson: string;
	    fitRating: string;
	    notes: string;
	    cvFile: string;
	    coverLetterFile: string;
	    source: string;
	    deadline: string;
	
	    static createFrom(source: any = {}) {
	        return new Application(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.row = source["row"];
	        this.date = source["date"];
	        this.company = source["company"];
	        this.sector = source["sector"];
	        this.role = source["role"];
	        this.roleType = source["roleType"];
	        this.channel = source["channel"];
	        this.status = source["status"];
	        this.contactPerson = source["contactPerson"];
	        this.fitRating = source["fitRating"];
	        this.notes = source["notes"];
	        this.cvFile = source["cvFile"];
	        this.coverLetterFile = source["coverLetterFile"];
	        this.source = source["source"];
	        this.deadline = source["deadline"];
	    }
	}
	export class QueueItem {
	    number: number;
	    done: boolean;
	    fit: string;
	    role: string;
	    company: string;
	    notes: string;
	    url: string;
	    outcome: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.number = source["number"];
	        this.done = source["done"];
	        this.fit = source["fit"];
	        this.role = source["role"];
	        this.company = source["company"];
	        this.notes = source["notes"];
	        this.url = source["url"];
	        this.outcome = source["outcome"];
	    }
	}

}

export namespace journal {
	
	export class PromptAnswer {
	    prompt: string;
	    answer: string;
	
	    static createFrom(source: any = {}) {
	        return new PromptAnswer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.prompt = source["prompt"];
	        this.answer = source["answer"];
	    }
	}
	export class Entry {
	    date: string;
	    prompts: PromptAnswer[];
	    body: string;
	    updatedAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Entry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.prompts = this.convertValues(source["prompts"], PromptAnswer);
	        this.body = source["body"];
	        this.updatedAt = this.convertValues(source["updatedAt"], time.Time);
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

export namespace main {
	
	export class JobsInfo {
	    dir: string;
	
	    static createFrom(source: any = {}) {
	        return new JobsInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	    }
	}

}

export namespace settings {
	
	export class Settings {
	    theme: string;
	    jobSearchDir: string;
	    lastReviewAt?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.jobSearchDir = source["jobSearchDir"];
	        this.lastReviewAt = this.convertValues(source["lastReviewAt"], time.Time);
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
	
	export class Project {
	    id: string;
	    title: string;
	    done: boolean;
	    createdAt: time.Time;
	    completedAt?: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.done = source["done"];
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
	export class Task {
	    id: string;
	    title: string;
	    done: boolean;
	    list: string;
	    contexts: string[];
	    projectId: string;
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
	        this.list = source["list"];
	        this.contexts = source["contexts"];
	        this.projectId = source["projectId"];
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

export namespace workouts {
	
	export class WorkoutLog {
	    index: number;
	    date: string;
	    amrapReps: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkoutLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.date = source["date"];
	        this.amrapReps = source["amrapReps"];
	    }
	}
	export class Cycle {
	    id: string;
	    number: number;
	    startDate: string;
	    trainingMax: Record<string, number>;
	    logs: WorkoutLog[];
	    createdAt: time.Time;
	
	    static createFrom(source: any = {}) {
	        return new Cycle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.number = source["number"];
	        this.startDate = source["startDate"];
	        this.trainingMax = source["trainingMax"];
	        this.logs = this.convertValues(source["logs"], WorkoutLog);
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

}

