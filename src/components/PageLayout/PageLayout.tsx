import styles from './pagelayout.module.scss'

import { Title, LeadParagraph } from '@statisticsnorway/ssb-component-library'

interface PageLayoutProps {
    title: string,
    description?: string,
    button?: JSX.Element,
    content?: JSX.Element
}

export default function PageLayout({title, description, button, content}: PageLayoutProps) {
    return (
        <>
            <div className={styles.title}>
                <Title size={1}>{title}</Title>
                {button}
            </div>
            <LeadParagraph>{description}</LeadParagraph>
            {content}
        </>
    )
}