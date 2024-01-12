import './PageLayout.scss'

import { Title, LeadParagraph } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
    title: string,
    description?: string,
    button?: JSX.Element,
    body?: JSX.Element
}

export default function PageLayout({title, description, button, body}: PageLayoutProps) {
    return (
        <div className="container">
            <div className="title-container">
                <Title size={1}>{title}</Title>
                {button}
            </div>
            <LeadParagraph>{description}</LeadParagraph>
            {body}
        </div>
    )
}