import './PageLayout.scss'
import { Title, Button } from '@statisticsnorway/ssb-component-library'

export function PageLayout({title, buttonText}: PageLayoutProps) {
    return (
        <div className="container">
            <div className="title-container">
                <Title size={1}>{title}</Title>
                {buttonText && <Button>{buttonText}</Button>}
            </div>
        </div>
    )
}

interface PageLayoutProps {
    title: string;
    buttonText?: string | ChildNode;
}