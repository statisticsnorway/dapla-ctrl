export interface RootObject {
    _embedded: EmbeddedTeams;
    _links: Links;
    count: number;
}

interface EmbeddedTeams {
    teams: Team[];
}

interface Team {
    uniformName: string;
    displayName: string;
    _links: TeamLinks;
}

interface TeamLinks {
    self: Link;
}

interface Link {
    href: string;
    templated?: boolean;
}

interface Links {
    self: Link;
}